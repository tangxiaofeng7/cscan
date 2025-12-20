import request from './request'

// 标签映射
export function getTagMappingList() {
  return request.post('/poc/tagmapping/list')
}

export function saveTagMapping(data) {
  return request.post('/poc/tagmapping/save', data)
}

export function deleteTagMapping(data) {
  return request.post('/poc/tagmapping/delete', data)
}

// 自定义POC
export function getCustomPocList(data) {
  return request.post('/poc/custom/list', data)
}

export function saveCustomPoc(data) {
  return request.post('/poc/custom/save', data)
}

// 批量导入自定义POC
export function batchImportCustomPoc(data) {
  return request.post('/poc/custom/batchImport', data)
}

export function deleteCustomPoc(data) {
  return request.post('/poc/custom/delete', data)
}

// 清空所有自定义POC
export function clearAllCustomPoc() {
  return request.post('/poc/custom/clearAll')
}

// Nuclei默认模板
export function getNucleiTemplateList(data) {
  return request.post('/poc/nuclei/templates', data)
}

export function getNucleiTemplateCategories() {
  return request.post('/poc/nuclei/categories')
}


// 同步Nuclei模板
export function syncNucleiTemplates(data = {}) {
  return request.post('/poc/nuclei/sync', data)
}

// 更新模板启用状态
export function updateTemplateEnabled(data) {
  return request.post('/poc/nuclei/updateEnabled', data)
}

// 获取模板详情
export function getNucleiTemplateDetail(data) {
  return request.post('/poc/nuclei/detail', data)
}

// 验证POC
export function validatePoc(data) {
  return request.post('/poc/custom/validate', data)
}

// 批量验证POC
export function batchValidatePoc(data) {
  return request.post('/poc/batchValidate', data)
}

// 查询POC验证结果
export function getPocValidationResult(data) {
  return request.post('/poc/queryResult', data)
}
