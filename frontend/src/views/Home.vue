<template>
  <v-layout
    :justify-center="!signedIn"
    :justify-start="signedIn"
    align-center
    column
  >
    <v-alert
      :type="alertType"
      :value="alertVisible"
      dismissable
    >
      {{ alertMessage }}
    </v-alert>
    <v-btn
      v-if="!signedIn"
      :disabled="!ready"
      color="primary"
      @click="signIn"
    >
      <v-icon
        v-if="ready"
        left
      >
        $vuetify.icon.signIn
      </v-icon>
      <v-progress-circular
        v-else
        color="primary"
        indeterminate
        :size="20"
      />
      <span>Sign In</span>
    </v-btn>
    <CertificateList v-else />
  </v-layout>
</template>

<script>
import CertificateList from '../components/CertificateList'
import { mapState, createNamespacedHelpers } from 'vuex'

const auth = createNamespacedHelpers('gAuth')

export default {
  name: 'Home',
  components: { CertificateList },
  computed: {
    ...mapState(['alertType', 'alertMessage', 'alertVisible']),
    ...auth.mapState(['ready', 'signedIn'])
  },
  methods: {
    ...auth.mapActions(['signIn'])
  }
}
</script>
