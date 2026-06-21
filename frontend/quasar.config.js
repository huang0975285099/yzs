/* eslint-env node */
const { configure } = require('quasar/wrappers')

module.exports = configure(function (/* ctx */) {
  return {
    boot: ['pinia', 'axios'],

    css: ['app.scss'],

    extras: [
      'material-icons'
    ],

    build: {
      target: {
        browser: ['es2019', 'edge88', 'firefox78', 'chrome87', 'safari13.1'],
        node: 'node20'
      },
      vueRouterMode: 'history',
      vitePlugins: []
    },

    devServer: {
      open: false,
      proxy: {
        '/api': {
          target: 'http://localhost:18881',
          changeOrigin: true
        },
        '/ws': {
          target: 'ws://localhost:18881',
          ws: true,
          changeOrigin: true
        }
      }
    },

    framework: {
      config: {
        notify: {
          position: 'top',
          timeout: 2500
        }
      },
      iconSet: 'material-icons',
      lang: 'zh-CN',
      plugins: ['Notify', 'Dialog', 'Loading', 'LocalStorage', 'SessionStorage']
    },

    animations: [],

    ssr: {
      pwa: false,
      prodPort: 3000,
      middlewares: ['render']
    },

    pwa: {
      workboxMode: 'generateSW',
      injectPwaMetaTags: true,
      swFilename: 'sw.js',
      manifestFilename: 'manifest.json',
      useCredentialsForManifestTag: false
    },

    capacitor: {
      hideSplashscreen: true
    },

    electron: {
      inspectPort: 5858,
      bundler: 'packager',
      packager: {},
      builder: {}
    }
  }
})
