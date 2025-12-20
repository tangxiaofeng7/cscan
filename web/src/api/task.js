import request from './request'

export function getTaskList(data) {
  return request.post('/task/list', data)
}

export function createTask(data) {
  return request.post('/task/create', data)
}

export function deleteTask(data) {
  return request.post('/task/delete', data)
}

export function batchDeleteTask(data) {
  return request.post('/task/batchDelete', data)
}

export function getTaskProfileList() {
  return request.post('/task/profile/list')
}

export function saveTaskProfile(data) {
  return request.post('/task/profile/save', data)
}

export function deleteTaskProfile(data) {
  return request.post('/task/profile/delete', data)
}

export function retryTask(data) {
  return request.post('/task/retry', data)
}

export function startTask(data) {
  return request.post('/task/start', data)
}

export function pauseTask(data) {
  return request.post('/task/pause', data)
}

export function resumeTask(data) {
  return request.post('/task/resume', data)
}

export function stopTask(data) {
  return request.post('/task/stop', data)
}
