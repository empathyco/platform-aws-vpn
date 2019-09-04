<template>
  <v-menu
    bottom
    left
    offset-y
  >
    <template v-slot:activator="{ on }">
      <v-btn
        fab
        outlined
        v-on="on"
      >
        <v-avatar>
          <v-icon v-if="!signedIn">
            $vuetify.icons.account
          </v-icon>
          <v-tooltip
            v-if="signedIn"
            bottom
          >
            <template v-slot:activator="{ on }">
              <img
                :alt="profile.name"
                :src="profile.image"
                v-on="on"
              >
            </template>
            <span>Logged in as <b>{{ profile.name }}</b> ({{ profile.email }})</span>
          </v-tooltip>
        </v-avatar>
      </v-btn>
    </template>

    <v-list>
      <v-list-item>
        <v-list-item-title
          v-if="!signedIn"
          @click="signIn"
        >
          Sign In
        </v-list-item-title>
        <v-list-item-title
          v-if="signedIn"
          @click="signOut"
        >
          Sign Out
        </v-list-item-title>
      </v-list-item>
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
