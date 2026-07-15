<script setup lang="ts">
  useHead({ title: 'Zalio ERP — Users' })

  const { users, fetchUsers, createUser, toggleActive } = useUsers()

  const showForm = ref(false)
  const saving = ref(false)
  const form = reactive({ name: '', email: '', password: '', role: 'staff' })

  onMounted(fetchUsers)

  async function submit() {
    saving.value = true
    try {
      await createUser({ ...form })
      form.name = ''
      form.email = ''
      form.password = ''
      form.role = 'staff'
      showForm.value = false
    } catch {
      // toast sudah ditangani di composable
    } finally {
      saving.value = false
    }
  }

  const fmt = (s: string) => new Date(s).toLocaleDateString('id-ID')
</script>

<template>
  <div class="page">
    <div class="page-body">
      <div class="page-header" style="display:flex; align-items:flex-start; justify-content:space-between; gap:12px">
        <div>
          <h1 class="page-title">Users</h1>
          <p class="page-subtitle">Kelola pengguna internal back-office &amp; hak aksesnya.</p>
        </div>
        <button class="btn-primary" @click="showForm = !showForm">
          {{ showForm ? 'Tutup' : '+ Tambah User' }}
        </button>
      </div>

      <form v-if="showForm" class="user-form" @submit.prevent="submit">
        <div class="user-form-grid">
          <div>
            <label class="login-label">Nama</label>
            <input v-model="form.name" class="text-input" placeholder="Nama lengkap" required>
          </div>
          <div>
            <label class="login-label">Email</label>
            <input v-model="form.email" type="email" class="text-input" placeholder="email@contoh.com" required>
          </div>
          <div>
            <label class="login-label">Password</label>
            <input v-model="form.password" type="password" class="text-input" placeholder="min. 6 karakter" required>
          </div>
          <div>
            <label class="login-label">Role</label>
            <select v-model="form.role" class="text-input">
              <option value="staff">staff</option>
              <option value="admin">admin</option>
            </select>
          </div>
        </div>
        <button class="btn-primary" :disabled="saving" type="submit" style="margin-top:14px; align-self:flex-start">
          {{ saving ? 'Menyimpan...' : 'Simpan User' }}
        </button>
      </form>

      <div class="table-card">
        <div class="table-scroll">
          <table class="data-table">
            <thead>
              <tr>
                <th style="width:60px">ID</th>
                <th>Nama</th>
                <th>Email</th>
                <th>Role</th>
                <th>Status</th>
                <th>Dibuat</th>
                <th class="text-center" style="width:120px">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in users" :key="u.id">
                <td class="font-semibold">{{ u.id }}</td>
                <td>{{ u.name }}</td>
                <td>{{ u.email }}</td>
                <td><span class="badge" :class="u.role === 'admin' ? 'badge-accent' : 'badge-muted'">{{ u.role }}</span></td>
                <td><span class="badge" :class="u.is_active ? 'badge-success' : 'badge-danger'">{{ u.is_active ? 'aktif' : 'nonaktif' }}</span></td>
                <td>{{ fmt(u.created_at) }}</td>
                <td class="text-center">
                  <button class="link-btn" @click="toggleActive(u)">
                    {{ u.is_active ? 'Nonaktifkan' : 'Aktifkan' }}
                  </button>
                </td>
              </tr>
              <tr v-if="!users.length">
                <td colspan="7" style="text-align:center; color:var(--text-muted); padding:28px">
                  Belum ada user
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
  .user-form {
    display: flex;
    flex-direction: column;
    background-color: var(--bg-surface);
    border: 1px solid var(--border-color);
    border-radius: 14px;
    padding: 20px;
    flex-shrink: 0;
  }
  .user-form-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 14px;
  }
  .link-btn {
    font-size: 13px;
    font-weight: 700;
    color: var(--accent);
    background: none;
    border: none;
    cursor: pointer;
  }
  .link-btn:hover {
    text-decoration: underline;
  }
  .badge-muted {
    background-color: var(--bg-muted);
    color: var(--text-muted);
  }
</style>
