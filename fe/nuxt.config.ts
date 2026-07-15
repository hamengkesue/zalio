export default defineNuxtConfig({
  modules: ['@nuxt/ui'],
  colorMode: { preference: 'light' },
  devtools: { enabled: false },
  future: { compatibilityVersion: 4 },
  devServer: { port: 3005 },

  css: ['~/assets/css/main.css'],

  app: {
    head: {
      title: 'Zalio ERP',
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=Nunito+Sans:wght@400;500;600;700;800&display=swap',
        },
      ],
    },
  },

  compatibilityDate: '2025-06-16',
})
