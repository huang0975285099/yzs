import axios from 'axios'
import router from '../router'

const http = axios.create({
  baseURL: '/api',
  timeout: 15000
})

http.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

http.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      router.push('/login')
    }
    return Promise.reject(error.response?.data || error)
  }
)

export const authApi = {
  login: (data) => http.post('/login', data),
  logout: () => http.post('/logout'),
  me: () => http.get('/me'),
  recordStart: () => http.post('/me/start'),
  recordSkip: () => http.post('/me/skip'),
  dailyStats: (params) => http.get('/me/daily-stats', { params }),
  saveHandledGoods: (data) => http.post('/me/handled-goods', data),
  listHandledGoods: (params) => http.get('/me/handled-goods', { params }),
  inspectStats: (params) => http.get('/me/inspect-stats', { params }),
  changePassword: (data) => http.post('/me/password', data)
}

export const userApi = {
  list: (params) => http.get('/users', params ? { params } : {}),
  create: (data) => http.post('/users', data),
  update: (id, data) => http.put(`/users/${id}`, data),
  delete: (id) => http.delete(`/users/${id}`)
}

export const teamApi = {
  list: () => http.get('/teams'),
  create: (data) => http.post('/teams', data),
  update: (id, data) => http.put(`/teams/${id}`, data),
  delete: (id) => http.delete(`/teams/${id}`)
}

export const tradeApi = {
  list: (params) => http.get('/trades', { params }),
  hourlyStats: (params) => http.get('/trades/hourly-stats', { params }),
  handle: (id, data) => http.post(`/trades/${id}/handle`, data),
  check: (id) => http.get(`/trades/${id}/check`),
  pend: (id, data) => http.post(`/trades/${id}/pend`, data),
  submit: (id, data) => http.post(`/trades/${id}/submit`, data),
  myHandled: (params) => http.get('/trades/my-handled', { params }),
  randomUnhandled: () => http.get('/trades/random-unhandled'),
  randomUninspected: () => http.get('/trades/random-uninspected'),
  lock: (id) => http.post(`/trades/${id}/lock`),
  unlock: (id) => http.post(`/trades/${id}/unlock`),
  detail: (id) => http.get(`/trades/${id}/detail`),
  branchProducts: (id, keyword) => http.get(`/trades/${id}/branch-products`, { params: { keyword } }),
  productPrice: (id, productId) => http.post(`/trades/${id}/product-price`, { productId }),
  inspect: (id, data) => http.post(`/trades/${id}/inspect`, data)
}

export const statsApi = {
  get: () => http.get('/stats'),
  operatorStats: () => http.get('/stats/operators'),
  operatorRecords: (params) => http.get('/stats/operator-records', { params }),
  daily: (params) => http.get('/stats/daily', { params }),
  operatorRange: (params) => http.get('/stats/operator-range', { params }),
  inspectExport: (params) => http.get('/stats/inspect-export', { params }),
  inspectorStats: () => http.get('/stats/inspectors'),
  inspectorRange: (params) => http.get('/stats/inspector-range', { params })
}

export const reviewApi = {
  list: (params) => http.get('/reviews', { params }),
  approve: (id) => http.post(`/reviews/${id}/approve`),
  remark: (id, data) => http.post(`/reviews/${id}/remark`, data)
}

export const goodsApi = {
  list: (params) => http.get('/goods', { params })
}

export const favoriteApi = {
  add: (data) => http.post('/favorites', data),
  remove: (goodsId) => http.delete(`/favorites/${goodsId}`),
  list: () => http.get('/favorites'),
  check: (goodsIds) => http.get('/favorites/check', { params: { goodsIds } })
}

export default http
