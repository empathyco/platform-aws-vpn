
const MomentLocalesPlugin = require('moment-locales-webpack-plugin')

module.exports = {
  configureWebpack: {
    plugins: [
      new MomentLocalesPlugin()
    ]
  },
  devServer: {
    proxy: {
      '/api': {
        target: 'https://replace-me-with-your-vpn-endpoint/api',
        changeOrigin: true,
        pathRewrite: {
          '^/api': ''
        }
      }
    }
  }
}
