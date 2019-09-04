<template>
  <v-card
    width="100%"
  >
    <v-card-title primary-title>
      My VPN Certificates
      <v-spacer />
      <v-spacer />
      <v-spacer />
      <v-text-field
        v-model="search"
        :disabled="!signedIn"
        prepend-icon="$vuetify.icons.search"
        label="Search"
        hint="Search for email, Serial Number or Key ID"
        single-line
        hide-details
      />
    </v-card-title>
    <v-card-actions>
      <v-btn
        :loading="isLoading"
        :disabled="!signedIn"
        outlined
        color="indigo"
        @click="doUpdateCerts"
      >
        <v-icon left>
          $vuetify.icons.refresh
        </v-icon>
        Refresh
      </v-btn>
      <v-btn
        :disabled="!signedIn"
        outlined
        color="green"
        @click="getCert"
      >
        <v-icon left>
          $vuetify.icons.certificate
        </v-icon>
        Request Certificate
      </v-btn>
      <v-spacer />
      <v-switch
        v-if="isAdmin"
        v-model="showAllUsers"
        label="Show all users"
        hide-details
        inset
        flat
      />
    </v-card-actions>
    <v-data-table
      :headers="headers"
      :items="certificates"
      :loading="isLoading"
      :search="search"
      item-key="certificates.serial"
      sort-by="notAfter"
      must-sort
      class="elevation-1"
    >
      <template v-slot:item.serial="{ item }">
        <v-tooltip top>
          <template v-slot:activator="{ on }">
            <pre v-on="on">{{ item.serial }}</pre>
          </template>
          Key ID: {{ item.keyId }}
        </v-tooltip>
      </template>

      <template v-slot:item.notBefore="{ item }">
        <v-tooltip top>
          <template v-slot:activator="{ on }">
            <span v-on="on">{{ item.notBefore | timeDistance }}</span>
          </template>
          {{ item.notBefore }}
        </v-tooltip>
      </template>

      <template v-slot:item.notAfter="{ item }">
        <v-tooltip top>
          <template v-slot:activator="{ on }">
            <span v-on="on">{{ item.notAfter | timeDistance }}</span>
          </template>
          {{ item.notAfter }}
        </v-tooltip>
      </template>

      <template v-slot:item.revoked="{ item }">
        <v-tooltip
          v-if="item.isRevoked"
          top
        >
          <template v-slot:activator="{ on }">
            <span v-on="on">{{ item.revoked | timeDistance }}</span>
          </template>
          {{ item.revoked }}
        </v-tooltip>
        <v-btn
          v-if="!item.isRevoked"
          outlined
          color="orange"
          @click="toRevoke = item"
        >
          <v-icon left>
            $vuetify.icons.revoke
          </v-icon>
          Revoke
        </v-btn>
      </template>

      <v-alert
        slot="no-data"
        :value="true"
        outlined
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
        outlined
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
        <v-card-text>Are you sure you want to revoke certificate<br><pre><b>{{ toRevoke.serial }}</b></pre></v-card-text>
        <v-card-actions>
          <v-spacer />
          <v-btn
            outlined
            @click="toRevoke = null"
          >
            <v-icon left>
              $vuetify.icons.cancel
            </v-icon>
            Cancel
          </v-btn>
          <v-btn
            color="red"
            outlined
            @click="revokeCert(toRevoke)"
          >
            <v-icon left>
              $vuetify.icons.revoke
            </v-icon>
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
        text
        @click.native="downloadedCert = false"
      >
        <v-icon>$vuetify.icons.close</v-icon>
      </v-btn>
    </v-snackbar>
  </v-card>
</template>

<script>
import { createNamespacedHelpers } from 'vuex'
import { compareAsc, distanceInWordsToNow } from 'date-fns'

const { mapState } = createNamespacedHelpers('certs')
const authMapState = createNamespacedHelpers('gAuth').mapState

export default {
  name: 'CertificateList',
  filters: {
    timeDistance: function (value) {
      return distanceInWordsToNow(value, { addSuffix: true })
    }
  },
  data () {
    return {
      showAllUsers: false,
      search: '',
      toRevoke: null,
      headers: [
        { text: 'Serial Number', value: 'serial', sortable: false, width: '20em' },
        { text: 'Subject', value: 'subject' },
        { text: 'Issued', value: 'notBefore', sort: compareAsc },
        { text: 'Expires', value: 'notAfter', sort: compareAsc },
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
