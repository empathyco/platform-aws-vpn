
const state = {
  ready: false,
  signedIn: false,
  id_token: null,
  profile: null
}

const mutations = {
  updateUser: (state, user) => {
    state.ready = true
    if (user.isSignedIn()) {
      state.signedIn = true

      let authData = user.getAuthResponse()
      state.id_token = authData.id_token

      let profile = user.getBasicProfile()
      state.profile = {
        id: profile.getId(),
        email: profile.getEmail(),
        name: profile.getName(),
        image: profile.getImageUrl()
      }
    } else {
      state.signedIn = false
      state.id_token = null
      state.profile = null
    }
  }
}

const actions = {
  initAuth (state) {
    if (!window.gapi) {
      console.error('gapi not found')
      return
    }
    window.gapi.load('auth2', () => {
      window.gapi.auth2.init({
        client_id: process.env.VUE_APP_GOOGLE_CLIENT_ID,
        hosted_domain: process.env.VUE_APP_GOOGLE_HOSTED_DOMAIN
      }).then(auth2 => {
        state.commit('updateUser', auth2.currentUser.get())
        auth2.currentUser.listen(user => {
          state.commit('updateUser', user)
        })
      }).catch(err => console.error('error on gauth init', err))
    })
  },
  async signIn () {
    try {
      let auth = window.gapi.auth2.getAuthInstance()
      await auth.signIn()
      console.debug('User signed in')
    } catch (e) {
      console.error(e)
    }
  },
  async signOut () {
    try {
      let auth = window.gapi.auth2.getAuthInstance()
      await auth.signOut()
      auth.disconnect()
      console.debug('User signed out')
    } catch (e) {
      console.error(e)
    }
  }
}

export default {
  namespaced: true,
  state,
  actions,
  mutations
}
