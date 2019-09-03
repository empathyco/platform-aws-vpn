<template>
  <v-app>
    <v-app-bar app>
      <v-spacer />
      <v-toolbar-title>{{ appName }}</v-toolbar-title>
      <v-spacer />
      <GoogleLogin />
    </v-app-bar>
    <v-content>
      <v-container
        align-content-center
        fill-height
        fluid
      >
        <router-view />
      </v-container>
    </v-content>
  </v-app>
</template>

<script>
import GoogleLogin from './components/GoogleLogin'

export default {
  components: {
    GoogleLogin
  },
  data: function () {
    return {
      appName: process.env.VUE_APP_NAME
    }
  }
}
</script>
