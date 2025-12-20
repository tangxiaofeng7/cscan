// MongoDB初始化脚本
db = db.getSiblingDB('cscan');

// 创建默认工作空间
var workspaceResult = db.workspace.insertOne({
    name: "默认工作空间",
    description: "系统默认工作空间",
    status: "enable",
    create_time: new Date(),
    update_time: new Date()
});

var defaultWorkspaceId = workspaceResult.insertedId.toString();

// 创建用户集合并插入默认管理员，关联默认工作空间
db.user.insertOne({
    username: "admin",
    password: "e10adc3949ba59abbe56e057f20f883e", // 123456的MD5
    role: "superadmin",
    status: "enable",
    workspace_ids: [defaultWorkspaceId],
    create_time: new Date(),
    update_time: new Date()
});

// 创建默认任务配置
db.task_profile.insertMany([
    {
        name: "快速扫描",
        description: "仅进行端口扫描和服务识别",
        config: JSON.stringify({
            portscan: { enable: true, ports: "21,22,80,443,3306,6379,8080" },
            fingerprint: { enable: true },
            pocscan: { enable: false }
        }),
        sort_number: 1,
        create_time: new Date(),
        update_time: new Date()
    },
    {
        name: "标准扫描",
        description: "端口扫描+服务识别+指纹识别",
        config: JSON.stringify({
            portscan: { enable: true, ports: "top1000" },
            fingerprint: { enable: true },
            pocscan: { enable: false }
        }),
        sort_number: 2,
        create_time: new Date(),
        update_time: new Date()
    },
    {
        name: "深度扫描",
        description: "全端口扫描+服务识别+指纹识别+漏洞扫描",
        config: JSON.stringify({
            portscan: { enable: true, ports: "1-65535" },
            fingerprint: { enable: true },
            pocscan: { enable: true, pocTypes: ["nuclei"] }
        }),
        sort_number: 3,
        create_time: new Date(),
        update_time: new Date()
    }
]);

// 创建索引
db.user.createIndex({ username: 1 }, { unique: true });
db.workspace.createIndex({ name: 1 });
db.task_profile.createIndex({ sort_number: 1 });

print("CSCAN MongoDB初始化完成");
