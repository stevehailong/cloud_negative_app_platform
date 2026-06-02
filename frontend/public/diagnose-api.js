// 在浏览器控制台中运行此代码来诊断 API 响应

console.log('=== Pipeline Runs API 诊断 ===');

// 1. 检查当前页面的 API 请求
async function checkAPI() {
    try {
        // 从浏览器发起请求（会自动带上认证信息）
        const response = await fetch('/api/v1/pipeline-runs?page=1&pageSize=10&pipeline_id=22');
        const data = await response.json();
        
        console.log('✅ API 响应状态:', response.status);
        console.log('📦 完整响应:', data);
        
        if (data.data && data.data.list) {
            console.log('📋 记录总数:', data.data.list.length);
            
            // 检查第一条记录
            if (data.data.list.length > 0) {
                const firstRun = data.data.list[0];
                console.log('🔍 第一条记录:', firstRun);
                console.log('  - runNo:', firstRun.runNo);
                console.log('  - status:', firstRun.status);
                console.log('  - imageUrl:', firstRun.imageUrl);
                console.log('  - imageUrl 是否存在:', 'imageUrl' in firstRun);
                console.log('  - imageUrl 值:', firstRun.imageUrl || '(空)');
                
                // 检查所有字段
                console.log('📝 所有字段:', Object.keys(firstRun));
                
                // 查找特定的 run
                const targetRun = data.data.list.find(r => r.runNo === 'book-service-ci-1780320879');
                if (targetRun) {
                    console.log('🎯 目标记录 (book-service-ci-1780320879):', targetRun);
                    console.log('   imageUrl:', targetRun.imageUrl);
                } else {
                    console.log('⚠️ 未找到 book-service-ci-1780320879 记录');
                }
            }
        }
        
        return data;
    } catch (error) {
        console.error('❌ 错误:', error);
    }
}

// 2. 运行检查
checkAPI().then(() => {
    console.log('=== 诊断完成 ===');
    console.log('提示: 如果 imageUrl 字段不存在或为空，说明后端没有正确返回数据');
});
