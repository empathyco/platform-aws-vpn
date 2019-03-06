import axios from 'axios'
import store from '../store'

let _api = axios.create({
  baseURL: '/api/client'
})

_api.interceptors.request.use((config) => {
  let idToken = store.state.gAuth.id_token
  if (idToken) {
    config.headers.common['Authorization'] = 'Bearer ' + idToken
  } else {
    delete config.headers.common['Authorization']
  }

  return config
})

export default {
  getCerts (allUsers) {
    let config = {}
    if (allUsers) {
      config.params = { 'all': allUsers }
    }
    return _api.get('/certificates', config)
  },
  newCert (publicKey) {
    return _api.put('/certificates', { 'publicKey': publicKey })
  },
  revokeCert (serial) {
    return _api.delete('/certificates/' + serial)
  }
}
