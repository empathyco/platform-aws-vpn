<template>
  <v-card>
    <v-card-title primary-title>
      <v-layout row>
        <v-flex xs4>
          <h1 class="headline mb-0">
            My VPN Certificates
          </h1>
        </v-flex>
        <v-flex xs4 />
        <v-flex xs2>
          <v-switch
            v-if="isAdmin"
            v-model="showAllUsers"
            label="Show all users"
            hide-details
          />
        </v-flex>
        <v-flex xs2>
          <v-text-field
            v-model="search"
            :disabled="!signedIn"
            append-icon="search"
            label="Search"
            hint="Search for email, Serial Number or Key ID"
          />
        </v-flex>
      </v-layout>
    </v-card-title>
    <v-card-actions>
      <v-layout row>
        <v-flex xs3>
          <v-btn
            :disabled="!signedIn"
            outline
            block
            color="green"
            @click="getCert"
          >
            <v-icon left>
              add_circle_outline
            </v-icon> Request Certificate
          </v-btn>
        </v-flex>
        <v-flex xs5 />
        <v-flex xs2>
        </v-flex>
        <v-flex xs2>
          <v-btn
            :loading="isLoading"
            :disabled="!signedIn"
            outline
            block
            color="indigo"
            @click="doUpdateCerts"
          >
            Refresh <v-icon right>
              refresh
            </v-icon>
          </v-btn>
        </v-flex>
      </v-layout>
    </v-card-actions>
    <v-data-table
      :headers="headers"
      :items="certificates"
      :loading="isLoading"
      :search="search"
      :pagination.sync="pagination"
      item-key="certificates.serial"
      must-sort
      class="elevation-1"
    >
      <template
        slot="items"
        slot-scope="props"
      >
        <td>
          <v-tooltip top>
            <pre slot="activator">{{ props.item.serial }}</pre>
            <span>Key ID: {{ props.item.keyId }}</span>
          </v-tooltip>
        </td>
        <td>{{ props.item.subject }}</td>
        <td>
          <v-tooltip top>
            <span slot="activator">
              {{ props.item.notBefore.fromNow() }}
            </span>
            <span>{{ props.item.notBefore.format('ll') }}</span>
          </v-tooltip>
        </td>
        <td>
          <v-tooltip top>
            <span slot="activator">
              {{ props.item.notAfter.fromNow() }}
            </span>
            <span>{{ props.item.notAfter.format('ll') }}</span>
          </v-tooltip>
        </td>
        <td v-if="props.item.isRevoked">
          <v-tooltip top>
            <span slot="activator">
              {{ props.item.revoked.fromNow() }}
            </span>
            <span>{{ props.item.revoked.format('ll') }}</span>
          </v-tooltip>
        </td>
        <td v-if="!props.item.isRevoked">
          <v-btn
            flat
            block
            color="orange"
            @click="toRevoke = props.item"
          >
            Revoke <v-icon right>
              block
            </v-icon>
          </v-btn>
        </td>
      </template>
      <v-alert
        slot="no-data"
        :value="true"
        outline
        color="info"
        icon="info"
      >
        <span v-if="!signedIn">
          Sign in to continue
        </span>
        <span v-else-if="isLoading">
          Loading. Please wait...
        </span>
        <span v-else>
          You have no certificates! Create one by clicking the button above
        </span>
      </v-alert>
      <v-alert
        slot="no-results"
        :value="true"
        outline
        color="info"
        icon="info"
      >
        Your search for "{{ search }}" found no results.
      </v-alert>
    </v-data-table>
    <v-dialog
      v-if="toRevoke != null"
      :value="toRevoke != null"
      persistent
      max-width="300"
      block
    >
      <v-card>
        <v-card-title class="headline">
          Revoke certificate?
        </v-card-title>
        <v-card-text>Are you sure you want to revoke certificate <pre>{{ toRevoke.serial }}</pre></v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            color="darken-1"
            flat
            @click="toRevoke = null"
          >
            Cancel
          </v-btn>
          <v-btn
            color="red darken-1"
            flat
            @click="revokeCert(toRevoke)"
          >
            Revoke
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
    <v-snackbar
      v-model="downloadedCert"
      :timeout="5000"
      bottom
    >
      Certificate downloaded
      <v-btn
        flat
        color="pink"
        @click.native="downloadedCert = false"
      >
        <v-icon>close</v-icon>
      </v-btn>
    </v-snackbar>
  </v-card>
</template>

<script>
import { createNamespacedHelpers } from 'vuex'
const { mapState } = createNamespacedHelpers('certs')
const authMapState = createNamespacedHelpers('gAuth').mapState

export default {
  name: 'CertificateList',
  data () {
    return {
      pagination: { 'sortBy': 'notAfter', 'descending': true },
      showAllUsers: false,
      search: '',
      toRevoke: null,
      headers: [
        { text: 'Serial Number', value: 'serial', sortable: false },
        { text: 'Subject', value: 'subject' },
        { text: 'Issued', value: 'notBefore' },
        { text: 'Expires', value: 'notAfter' },
        { text: 'Revoked', value: 'revoked', sortable: false }
      ]
    }
  },
  computed: {
    ...mapState(['certificates', 'isLoading', 'isAdmin', 'downloadedCert']),
    ...authMapState(['signedIn'])
  },
  watch: {
    signedIn () {
      this.doUpdateCerts()
    },
    showAllUsers () {
      this.doUpdateCerts()
    }
  },
  mounted () {
    this.doUpdateCerts()
  },
  methods: {
    async doUpdateCerts () {
      if (this.signedIn) {
        await this.$store.dispatch('certs/updateCerts', this.showAllUsers)
      } else {
        await this.$store.dispatch('certs/clear')
      }
    },
    async getCert () {
      await this.$store.dispatch('certs/getCert')
      this.doUpdateCerts()
    },
    async revokeCert (cert) {
      await this.$store.dispatch('certs/revokeCert', cert.serial)
      this.toRevoke = null
      this.doUpdateCerts()
    }
  }
}
</script>

<style scoped>
.mono {
  font-family: monospace;
}
</style>
