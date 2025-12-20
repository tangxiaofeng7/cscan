import request from './request'

export function login(data) {
  return request.post('/login', data)
}

export function getUserList(data) {
  return request.post('/user/list', data)
}
