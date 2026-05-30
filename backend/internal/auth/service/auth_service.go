package service

import (
	"errors"
	"fmt"
	"my-cloud/internal/auth/repository"
	"my-cloud/internal/common/model"
	"my-cloud/pkg/jwt"
	"my-cloud/pkg/security"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo       *repository.UserRepository
	roleRepo       *repository.RoleRepository
	jwtSecret      string
	settings       *security.SettingsLoader
	loginLimiter   *security.LoginLimiter
}

func NewAuthService(userRepo *repository.UserRepository, roleRepo *repository.RoleRepository, jwtSecret string, settings *security.SettingsLoader) *AuthService {
	jwt.InitJWT(jwtSecret)
	return &AuthService{
		userRepo:     userRepo,
		roleRepo:     roleRepo,
		jwtSecret:    jwtSecret,
		settings:     settings,
		loginLimiter: security.NewLoginLimiter(),
	}
}

// Login 用户登录
func (s *AuthService) Login(username, password, ip string) (string, string, *model.User, error) {
	sec := s.settings.Get()

	// 检查是否被锁定
	if locked, minutes := s.loginLimiter.IsLocked(username, sec); locked {
		return "", "", nil, fmt.Errorf("账户已被锁定，请%d分钟后重试", minutes)
	}

	// 获取用户
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		s.loginLimiter.RecordFailure(username, sec)
		return "", "", nil, errors.New("用户名或密码错误")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		locked, minutes := s.loginLimiter.RecordFailure(username, sec)
		if locked {
			return "", "", nil, fmt.Errorf("密码错误次数过多，账户已锁定%d分钟", minutes)
		}
		return "", "", nil, errors.New("用户名或密码错误")
	}

	// 检查用户状态
	if user.Status != 1 {
		return "", "", nil, errors.New("用户已被禁用")
	}

	// 登录成功，清除失败记录
	s.loginLimiter.RecordSuccess(username)

	// 更新最后登录信息
	loginTime := time.Now().Unix()
	_ = s.userRepo.UpdateLastLogin(user.ID, loginTime, ip)

	// 生成access token（使用系统设置的session超时时间）
	expireSeconds := sec.SessionTimeout * 60
	accessToken, err := jwt.GenerateToken(user.ID, user.Username, expireSeconds)
	if err != nil {
		return "", "", nil, err
	}

	// 生成refresh token (7天)
	refreshToken, err := jwt.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return "", "", nil, err
	}

	return accessToken, refreshToken, user, nil
}

// Register 用户注册
func (s *AuthService) Register(user *model.User) error {
	// 检查用户名是否存在
	if existUser, _ := s.userRepo.GetByUsername(user.Username); existUser != nil {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否存在
	if user.Email != "" {
		if existUser, _ := s.userRepo.GetByEmail(user.Email); existUser != nil {
			return errors.New("邮箱已存在")
		}
	}

	// 验证密码复杂度
	if err := security.ValidatePassword(user.Password, s.settings.Get()); err != nil {
		return err
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// 创建用户
	if err := s.userRepo.Create(user); err != nil {
		return err
	}

	// 自动分配默认角色GUEST (ID=5)
	guestRole, err := s.roleRepo.GetByCode("GUEST")
	if err == nil && guestRole != nil {
		_ = s.roleRepo.AssignRoleToUser(user.ID, guestRole.ID)
	}

	return nil
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(userID uint) (*model.User, []*model.Role, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, nil, err
	}

	roles, err := s.roleRepo.GetUserRoles(userID)
	if err != nil {
		return nil, nil, err
	}

	return user, roles, nil
}

// UpdatePassword 修改密码
func (s *AuthService) UpdatePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("原密码错误")
	}

	// 验证新密码复杂度
	if err := security.ValidatePassword(newPassword, s.settings.Get()); err != nil {
		return err
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.userRepo.Update(user)
}

// UpdateProfile 更新用户资料
func (s *AuthService) UpdateProfile(userID uint, updates map[string]interface{}) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// 更新允许的字段
	if realName, ok := updates["realName"].(string); ok {
		user.RealName = realName
	}
	if phone, ok := updates["phone"].(string); ok {
		user.Phone = phone
	}
	if avatar, ok := updates["avatar"].(string); ok {
		user.Avatar = avatar
	}
	if department, ok := updates["department"].(string); ok {
		user.Department = department
	}
	if position, ok := updates["position"].(string); ok {
		user.Position = position
	}

	return s.userRepo.Update(user)
}

// GenerateToken 生成token
func (s *AuthService) GenerateToken(userID uint, username string) (string, error) {
	expireSeconds := s.settings.Get().SessionTimeout * 60
	return jwt.GenerateToken(userID, username, expireSeconds)
}

// GetUserList 获取用户列表
func (s *AuthService) GetUserList(page, pageSize int, keyword string) ([]*model.User, int64, error) {
	return s.userRepo.List(page, pageSize, keyword)
}

// AssignRoles 为用户分配角色
func (s *AuthService) AssignRoles(userID uint, roleIDs []uint) error {
	// 检查用户是否存在
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	// 先删除用户的所有角色
	if err := s.roleRepo.DeleteUserRoles(userID); err != nil {
		return err
	}

	// 分配新角色
	for _, roleID := range roleIDs {
		if err := s.roleRepo.AssignRole(userID, roleID); err != nil {
			return err
		}
	}

	return nil
}

// UpdateUserStatus 更新用户状态
func (s *AuthService) UpdateUserStatus(userID uint, status int) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	user.Status = status
	return s.userRepo.Update(user)
}

// GetAllRoles 获取所有角色
func (s *AuthService) GetAllRoles() ([]*model.Role, error) {
	return s.roleRepo.GetAllRoles()
}

// GetUserRoles 获取用户角色列表
func (s *AuthService) GetUserRoles(userID uint) ([]*model.Role, error) {
	return s.roleRepo.GetUserRoles(userID)
}

// CreateUser 管理员创建用户
func (s *AuthService) CreateUser(user *model.User) error {
	// 检查用户名是否存在
	if existUser, _ := s.userRepo.GetByUsername(user.Username); existUser != nil {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否存在
	if user.Email != "" {
		if existUser, _ := s.userRepo.GetByEmail(user.Email); existUser != nil {
			return errors.New("邮箱已存在")
		}
	}

	// 如果提供了密码则验证并加密，否则使用默认密码
	if user.Password == "" {
		user.Password = "Abc123456"
	}
	if err := security.ValidatePassword(user.Password, s.settings.Get()); err != nil {
		return err
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)
	user.Status = 1

	return s.userRepo.Create(user)
}

// DeleteUser 删除用户
func (s *AuthService) DeleteUser(userID uint) error {
	_, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}
	// 先删除用户角色关联
	_ = s.roleRepo.DeleteUserRoles(userID)
	return s.userRepo.Delete(userID)
}

// UpdateUser 更新用户信息（管理员）
func (s *AuthService) UpdateUser(userID uint, updates map[string]interface{}) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return errors.New("用户不存在")
	}

	if realName, ok := updates["realName"].(string); ok {
		user.RealName = realName
	}
	if phone, ok := updates["phone"].(string); ok {
		user.Phone = phone
	}
	if email, ok := updates["email"].(string); ok {
		user.Email = email
	}
	if department, ok := updates["department"].(string); ok {
		user.Department = department
	}
	if position, ok := updates["position"].(string); ok {
		user.Position = position
	}

	return s.userRepo.Update(user)
}

// CreateRole 创建角色
func (s *AuthService) CreateRole(role *model.Role) error {
	return s.roleRepo.Create(role)
}

// RefreshToken 刷新token
func (s *AuthService) RefreshToken(refreshToken string) (string, string, error) {
	// 解析refresh token (不验证过期时间)
	claims, err := jwt.ParseTokenWithoutValidation(refreshToken)
	if err != nil {
		return "", "", errors.New("无效的refresh token")
	}

	// 检查用户是否仍然存在且状态正常
	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil {
		return "", "", errors.New("用户不存在")
	}

	if user.Status != 1 {
		return "", "", errors.New("用户已被禁用")
	}

	// 生成新的access token（使用系统设置的session超时）
	expireSeconds := s.settings.Get().SessionTimeout * 60
	newAccessToken, err := jwt.GenerateToken(user.ID, user.Username, expireSeconds)
	if err != nil {
		return "", "", err
	}

	// 生成新的refresh token (7天)
	newRefreshToken, err := jwt.GenerateRefreshToken(user.ID, user.Username)
	if err != nil {
		return "", "", err
	}

	return newAccessToken, newRefreshToken, nil
}
