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

// 创建HTTP服务映射集合和索引
db.http_service_mapping.createIndex({ service_name: 1 }, { unique: true });

// 初始化HTTP服务映射数据
var now = new Date();

// HTTP服务列表
var httpServices = [
    { service_name: "http", is_http: true, description: "标准HTTP服务", enabled: true, create_time: now, update_time: now },
    { service_name: "https", is_http: true, description: "标准HTTPS服务", enabled: true, create_time: now, update_time: now },
    { service_name: "http-proxy", is_http: true, description: "HTTP代理服务", enabled: true, create_time: now, update_time: now },
    { service_name: "https-alt", is_http: true, description: "HTTPS备用端口", enabled: true, create_time: now, update_time: now },
    { service_name: "http-alt", is_http: true, description: "HTTP备用端口", enabled: true, create_time: now, update_time: now },
    { service_name: "ajp12", is_http: true, description: "Apache JServ Protocol", enabled: true, create_time: now, update_time: now },
    { service_name: "esmagent", is_http: true, description: "Enterprise Security Manager Agent", enabled: true, create_time: now, update_time: now }
];

// 非HTTP服务列表
var nonHttpServices = [
    { service_name: "ssh", is_http: false, description: "SSH远程登录", enabled: true, create_time: now, update_time: now },
    { service_name: "ftp", is_http: false, description: "FTP文件传输", enabled: true, create_time: now, update_time: now },
    { service_name: "smtp", is_http: false, description: "邮件发送", enabled: true, create_time: now, update_time: now },
    { service_name: "pop3", is_http: false, description: "邮件接收", enabled: true, create_time: now, update_time: now },
    { service_name: "imap", is_http: false, description: "邮件访问", enabled: true, create_time: now, update_time: now },
    { service_name: "mysql", is_http: false, description: "MySQL数据库", enabled: true, create_time: now, update_time: now },
    { service_name: "mssql", is_http: false, description: "SQL Server数据库", enabled: true, create_time: now, update_time: now },
    { service_name: "oracle", is_http: false, description: "Oracle数据库", enabled: true, create_time: now, update_time: now },
    { service_name: "postgresql", is_http: false, description: "PostgreSQL数据库", enabled: true, create_time: now, update_time: now },
    { service_name: "redis", is_http: false, description: "Redis缓存", enabled: true, create_time: now, update_time: now },
    { service_name: "mongodb", is_http: false, description: "MongoDB数据库", enabled: true, create_time: now, update_time: now },
    { service_name: "memcached", is_http: false, description: "Memcached缓存", enabled: true, create_time: now, update_time: now },
    { service_name: "elasticsearch", is_http: false, description: "Elasticsearch搜索", enabled: true, create_time: now, update_time: now },
    { service_name: "dns", is_http: false, description: "DNS服务", enabled: true, create_time: now, update_time: now },
    { service_name: "snmp", is_http: false, description: "网络管理", enabled: true, create_time: now, update_time: now },
    { service_name: "ldap", is_http: false, description: "目录服务", enabled: true, create_time: now, update_time: now },
    { service_name: "smb", is_http: false, description: "文件共享", enabled: true, create_time: now, update_time: now },
    { service_name: "netbios", is_http: false, description: "NetBIOS服务", enabled: true, create_time: now, update_time: now },
    { service_name: "rdp", is_http: false, description: "远程桌面", enabled: true, create_time: now, update_time: now },
    { service_name: "vnc", is_http: false, description: "VNC远程桌面", enabled: true, create_time: now, update_time: now },
    { service_name: "telnet", is_http: false, description: "Telnet远程登录", enabled: true, create_time: now, update_time: now },
    { service_name: "rpc", is_http: false, description: "RPC服务", enabled: true, create_time: now, update_time: now },
    { service_name: "ntp", is_http: false, description: "时间同步", enabled: true, create_time: now, update_time: now },
    { service_name: "tftp", is_http: false, description: "简单文件传输", enabled: true, create_time: now, update_time: now },
    { service_name: "sip", is_http: false, description: "SIP通信", enabled: true, create_time: now, update_time: now },
    { service_name: "rtsp", is_http: false, description: "流媒体", enabled: true, create_time: now, update_time: now }
];

// 插入HTTP服务映射数据
db.http_service_mapping.insertMany(httpServices.concat(nonHttpServices));

print("CSCAN MongoDB初始化完成");
