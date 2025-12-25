/**
 * 目标格式校验工具
 */

/**
 * 校验单个目标
 * @param {string} target - 目标字符串
 * @returns {string|null} - 错误信息，如果有效则返回 null
 */
export function validateSingleTarget(target) {
  target = target.trim()
  if (!target || target.startsWith('#')) {
    return null // 空行或注释行
  }

  // 去除可能的端口部分
  let host = target
  const lastColon = target.lastIndexOf(':')
  if (lastColon !== -1) {
    const portStr = target.substring(lastColon + 1)
    const port = parseInt(portStr, 10)
    if (!isNaN(port) && port > 0 && port <= 65535) {
      host = target.substring(0, lastColon)
    }
  }

  // CIDR 格式
  if (host.includes('/')) {
    return validateCIDR(host)
  }

  // IP 范围格式
  if (host.includes('-')) {
    const parts = host.split('-')
    if (parts.length === 2 && isValidIP(parts[0].trim())) {
      return validateIPRange(host)
    }
    // 可能是域名中包含连字符
  }

  // 单个 IP
  if (isValidIP(host)) {
    return null
  }

  // 域名格式
  if (isValidDomain(host)) {
    return null
  }

  return '无效的目标格式，请输入有效的IP、CIDR、IP范围或域名'
}

/**
 * 校验 CIDR 格式
 * @param {string} cidr - CIDR 字符串
 * @returns {string|null} - 错误信息
 */
function validateCIDR(cidr) {
  const parts = cidr.split('/')
  if (parts.length !== 2) {
    return '无效的CIDR格式'
  }

  const ipPart = parts[0]
  const maskPart = parts[1]

  // 检查掩码
  const mask = parseInt(maskPart, 10)
  if (isNaN(mask) || mask < 0 || mask > 32) {
    return `无效的子网掩码: ${maskPart}`
  }

  // 检查 IP 部分是否完整
  const octets = ipPart.split('.')
  if (octets.length !== 4) {
    const suggestion = suggestCIDRFix(ipPart, maskPart)
    return `IP地址不完整，缺少${4 - octets.length}个八位组。正确格式示例: ${suggestion}`
  }

  // 验证每个八位组
  for (let i = 0; i < octets.length; i++) {
    const val = parseInt(octets[i], 10)
    if (isNaN(val) || val < 0 || val > 255) {
      return `第${i + 1}个八位组 '${octets[i]}' 无效，应为0-255之间的数字`
    }
  }

  return null
}

/**
 * 提供 CIDR 修复建议
 */
function suggestCIDRFix(ipPart, maskPart) {
  const octets = ipPart.split('.')
  while (octets.length < 4) {
    octets.push('0')
  }
  return octets.join('.') + '/' + maskPart
}

/**
 * 校验 IP 范围格式
 */
function validateIPRange(ipRange) {
  const parts = ipRange.split('-')
  if (parts.length !== 2) {
    return '无效的IP范围格式'
  }

  const startIP = parts[0].trim()
  const endIP = parts[1].trim()

  if (!isValidIP(startIP)) {
    return `起始IP '${startIP}' 无效`
  }
  if (!isValidIP(endIP)) {
    return `结束IP '${endIP}' 无效`
  }

  // 检查起始IP是否小于等于结束IP
  const start = startIP.split('.').map(Number)
  const end = endIP.split('.').map(Number)

  for (let i = 0; i < 4; i++) {
    if (start[i] > end[i]) {
      return '起始IP不能大于结束IP'
    }
    if (start[i] < end[i]) {
      break
    }
  }

  return null
}

/**
 * 检查是否是有效的 IPv4 地址
 */
function isValidIP(ip) {
  const parts = ip.split('.')
  if (parts.length !== 4) return false
  
  for (const part of parts) {
    const num = parseInt(part, 10)
    if (isNaN(num) || num < 0 || num > 255 || part !== String(num)) {
      return false
    }
  }
  return true
}

/**
 * 检查是否是有效的域名
 */
function isValidDomain(domain) {
  const domainRegex = /^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$/
  return domainRegex.test(domain)
}

/**
 * 校验多行目标
 * @param {string} targets - 多行目标字符串
 * @returns {Array<{line: number, target: string, message: string}>} - 错误列表
 */
export function validateTargets(targets) {
  const errors = []
  const lines = targets.split('\n')

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i].trim()
    if (!line || line.startsWith('#')) {
      continue
    }

    const error = validateSingleTarget(line)
    if (error) {
      errors.push({
        line: i + 1,
        target: line,
        message: error
      })
    }
  }

  return errors
}

/**
 * 格式化校验错误为用户友好的消息
 * @param {Array} errors - 错误列表
 * @returns {string} - 格式化的错误消息
 */
export function formatValidationErrors(errors) {
  if (errors.length === 0) return ''

  if (errors.length === 1) {
    const e = errors[0]
    return `第${e.line}行 '${e.target}': ${e.message}`
  }

  const messages = errors.map(e => `第${e.line}行 '${e.target}': ${e.message}`)
  return `发现${errors.length}个目标格式错误:\n${messages.join('\n')}`
}
