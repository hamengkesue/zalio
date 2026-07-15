// Penjaga rute global: setiap halaman butuh login, KECUALI /login.
// Berjalan sebelum halaman dirender.
export default defineNuxtRouteMiddleware((to) => {
  const token = useCookie('zalio_token')

  if (to.path === '/login') {
    // Sudah login? Jangan biarkan lihat halaman login lagi.
    if (token.value) return navigateTo('/')
    return
  }

  if (!token.value) {
    return navigateTo('/login')
  }
})
