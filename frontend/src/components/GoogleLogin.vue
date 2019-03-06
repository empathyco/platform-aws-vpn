<template>
  <v-menu
    bottom
    left
    offset-y
  >
    <v-avatar slot="activator">
      <v-icon v-if="!signedIn">
        account_circle
      </v-icon>
      <v-tooltip
        v-if="signedIn"
        bottom
      >
        <img
          slot="activator"
          :alt="profile.name"
          :src="profile.image"
        >
        <span>Logged in as <b>{{ profile.name }}</b> ({{ profile.email }})</span>
      </v-tooltip>
    </v-avatar>
    <v-list>
      <v-list-tile>
        <v-list-tile-title
          v-if="!signedIn"
          @click="signIn"
        >
          Sign In
        </v-list-tile-title>
        <v-list-tile-title
          v-if="signedIn"
          @click="signOut"
        >
          Sign Out
        </v-list-tile-title>
      </v-list-tile>
    </v-list>
  </v-menu>
</template>

<script>
import { createNamespacedHelpers } from 'vuex'
const { mapState, mapActions } = createNamespacedHelpers('gAuth')

export default {
  name: 'GoogleLogin',
  computed: {
    ...mapState(['id_token', 'profile', 'signedIn'])
  },
  mounted: function () {
    this.initAuth()
  },
  methods: {
    ...mapActions(['initAuth', 'signIn', 'signOut'])
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
  .jwt-token {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    width: 100em;
  }
</style>
