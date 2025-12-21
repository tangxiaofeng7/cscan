import request from './request'

// 获取指纹列表
export function getFingerprintList(data) {
  return request.post('/fingerprint/list', data)
}

// 保存指纹
export function saveFingerprint(data) {
  return request.post('/fingerprint/save', data)
}

// 删除指纹
export function deleteFingerprint(data) {
  return request.post('/fingerprint/delete', data)
}

// 获取指纹分类和统计
export function getFingerprintCategories() {
  return request.post('/fingerprint/categories')
}

// 同步指纹
export function syncFingerprints(data = {}) {
  return request.post('/fingerprint/sync', data)
}

// 更新指纹启用状态
export function updateFingerprintEnabled(data) {
  return request.post('/fingerprint/updateEnabled', data)
}

// 导入指纹
export function importFingerprints(data) {
  return request.post('/fingerprint/import', data)
}

// 清空自定义指纹
export function clearCustomFingerprints(data = {}) {
  return request.post('/fingerprint/clearCustom', data)
}

// 验证指纹
export function validateFingerprint(data) {
  return request.post('/fingerprint/validate', data)
}

// 批量验证指纹
export function batchValidateFingerprints(data) {
  return request.post('/fingerprint/batchValidate', data)
}


// ==================== HTTP服务映射 API ====================

// 获取HTTP服务映射列表
export function getHttpServiceMappingList(data = {}) {
  return request.post('/fingerprint/httpservice/list', data)
}

// 保存HTTP服务映射
export function saveHttpServiceMapping(data) {
  return request.post('/fingerprint/httpservice/save', data)
}

// 删除HTTP服务映射
export function deleteHttpServiceMapping(data) {
  return request.post('/fingerprint/httpservice/delete', data)
}
