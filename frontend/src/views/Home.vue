<template>
  <div class="home">
    <v-alert
      :type="alertType"
      :value="alertVisible"
      dismissable
    >
      {{ alertMessage }}
    </v-alert>
    <v-card v-if="!signedIn">
      <v-card-text class="title pa-5">
        <p
          v-if="!ready"
          class="text-xs-center"
        >
          <v-progress-circular
            indeterminate
            color="primary"
            :size="20"
          /> Loading...
        </p>
        <div
          v-else-if="!signedIn"
          class="text-xs-center"
        >
          <v-btn
            color="primary"
            @click="signIn"
          >
            <v-icon left>
              security
            </v-icon> Sign In
          </v-btn>
        </div>
      </v-card-text>
    </v-card>
    <CertificateList v-else />
  </div>
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
