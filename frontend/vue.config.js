
const BundleAnalyzerPlugin = require('webpack-bundle-analyzer').BundleAnalyzerPlugin

module.exports = {
  configureWebpack: {
    plugins: [
      new BundleAnalyzerPlugin({ analyzerMode: 'static', openAnalyzer: false })
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
