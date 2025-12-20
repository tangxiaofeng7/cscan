import request from './request'

export function getAssetList(data) {
  return request.post('/asset/list', data)
}

export function getAssetStat(data = {}) {
  return request.post('/asset/stat', data)
}

export function deleteAsset(data) {
  return request.post('/asset/delete', data)
}

export function batchDeleteAsset(data) {
  return request.post('/asset/batchDelete', data)
}

export function getAssetHistory(data) {
  return request.post('/asset/history', data)
}
