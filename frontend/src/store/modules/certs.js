
import moment from 'moment'
import api from '../../api'

import { downloadFile } from '../../utils/download'
import { newKeyPair, exportKeys } from '../../utils/crypto'

const state = {
  certificates: [],
  isAdmin: false,
  isLoading: false,
  downloadedCert: false
}

const getters = {}

const actions = {
  async updateCerts ({ state, commit }, showAllUsers) {
    await commit('setLoading', true)
    try {
      // TODO: error handling
      const r = await api.getCerts(showAllUsers)
      r.data.certs = r.data.certs.map(c => {
        c.notBefore = moment(c.notBefore)
        c.notAfter = moment(c.notAfter)

        c.isRevoked = c.revoked !== undefined
        if (c.isRevoked) {
          c.revoked = moment(c.revoked)
        }
        return c
      })
      await commit('setCertResponse', r.data)
    } catch (e) {
      console.error(e)
    } finally {
      await commit('setLoading', false)
    }
  },
  async revokeCert ({ state, commit }, serial) {
    try {
      // TODO: error handling
      await api.revokeCert(serial)
    } catch (e) {
      console.error(e)
    }
  },
  async getCert ({ state, commit }) {
    try {
      let keyPair = await newKeyPair()
      let jwks = await exportKeys(keyPair)

      let res = await api.newCert(jwks.public)
      let config = res.data.replace('%PRIVATEKEY%', jwks.private)
      let filename = res.headers['x-vpn-filename'] || `${process.env.VUE_APP_NAME}.ovpn`

      await downloadFile(filename, config, res.headers['content-type'])
      await commit('setDownloadedCert', true)
    } catch (e) {
      console.error(e)
    }
  },
  async clear ({ state, commit }) {
    await commit('setCertResponse', { certs: [], isAdmin: false })
    await commit('setLoading', false)
  }
}

const mutations = {
  setCertResponse (state, data) {
    state.certificates = data.certs
    state.isAdmin = data.isAdmin
  },
  setDownloadedCert (state, downloadedCert) {
    state.downloadedCert = downloadedCert
  },
  setLoading (state, isLoading) {
    state.isLoading = isLoading
  }
}

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations
}
