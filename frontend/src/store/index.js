import Vue from 'vue'
import Vuex from 'vuex'

import gauth from './modules/gauth'
import certs from './modules/certs'

Vue.use(Vuex)

const store = new Vuex.Store({
  strict: process.env.NODE_ENV !== 'production',
  modules: {
    gAuth: gauth,
    certs: certs
  },
  state: {
    alertType: 'error',
    alertMessage: '',
    alertVisible: false
  },
  mutations: {
    alert: function (state, { alertType, alertMessage }) {
      state.alertType = alertType
      state.alertMessage = alertMessage
      state.alertVisible = true
    }
  },
  actions: {
    async displayAlert ({ commit }, alertType, alertMessage) {
      await commit('alert', { alertType, alertMessage })
    }
  }
})

export default store
