<template>
  <v-app>
    <v-toolbar app>
      <v-spacer />
      <v-toolbar-title>{{ appName }}</v-toolbar-title>
      <v-spacer />
      <v-toolbar-items>
        <GoogleLogin />
      </v-toolbar-items>
    </v-toolbar>
    <v-content>
      <v-container fixed>
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

<style>
  body {
    height: 100%;
  }
  .card--flex-toolbar {
    margin-top: -65px;
  }
</style>
