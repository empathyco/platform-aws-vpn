import Vue from 'vue'
import Vuetify from 'vuetify/lib'
import {
  mdiAccountCircleOutline,
  mdiCancel,
  mdiCertificate, mdiClose,
  mdiRefresh,
  mdiShieldAccountOutline, mdiShieldOffOutline,
  mdiTableSearch
} from '@mdi/js'

Vue.use(Vuetify)

export default new Vuetify({
  icons: {
    iconfont: 'mdiSvg',
    values: {
      account: mdiAccountCircleOutline,
      cancel: mdiCancel,
      certificate: mdiCertificate,
      close: mdiClose,
      refresh: mdiRefresh,
      revoke: mdiShieldOffOutline,
      search: mdiTableSearch,
      signIn: mdiShieldAccountOutline
    }
  }
})
