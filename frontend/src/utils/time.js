/**
 * 统一时间格式化工具
 * 格式: YYYY-MM-DD HH:mm:ss
 */

/**
 * 格式化时间为 年-月-日 时:分:秒
 * @param {string|Date|number} time - 时间值
 * @returns {string} 格式化后的时间字符串
 */
export function formatTime(time) {
  if (!time) return '-'
  const date = new Date(time)
  if (isNaN(date.getTime())) return '-'
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  const seconds = String(date.getSeconds()).padStart(2, '0')
  return `${year}年${month}月${day}日 ${hours}时${minutes}分${seconds}秒`
}

/**
 * 格式化时间为 年-月-日
 * @param {string|Date|number} time - 时间值
 * @returns {string} 格式化后的日期字符串
 */
export function formatDate(time) {
  if (!time) return '-'
  const date = new Date(time)
  if (isNaN(date.getTime())) return '-'
  const year = date.getFullYear()
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const day = String(date.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

/**
 * 格式化耗时（秒 → 中文可读格式）
 * @param {number} seconds - 秒数
 * @returns {string} 格式化后的耗时字符串
 */
export function formatDuration(seconds) {
  if (!seconds) return '-'
  const minutes = Math.floor(seconds / 60)
  const secs = seconds % 60
  return minutes > 0 ? `${minutes}分${secs}秒` : `${secs}秒`
}
