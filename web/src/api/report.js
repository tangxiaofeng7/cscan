import request from './request'

// 获取报告详情
export function getReportDetail(data) {
  return request({
    url: '/report/detail',
    method: 'post',
    data
  })
}

// 导出报告
export function exportReport(data) {
  return request({
    url: '/report/export',
    method: 'post',
    data,
    responseType: 'blob'
  })
}
