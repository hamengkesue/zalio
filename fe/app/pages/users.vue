<script setup lang="ts">
  useHead({ title: 'Zalio ERP — Users' })

  const { users, fetchUsers, createUser, toggleActive } = useUsers()

  const showForm = ref(false)
  const saving = ref(false)
  const form = reactive({ name: '', username: '', email: '', password: '', role: 'staff' })

  onMounted(fetchUsers)

  async function submit() {
    saving.value = true
    try {
      await createUser({ ...form })
      form.name = ''
      form.username = ''
      form.email = ''
      form.password = ''
      form.role = 'staff'
      showForm.value = false
    } catch {
      // toast handled in composable
    } finally {
      saving.value = false
    }
  }

  const fmt = (s: string) => new Date(s).toLocaleDateString('en-GB')
</script>

<template>
  <div class="page">
    <div class="page-body">
      <div class="page-header" style="display:flex; align-items:flex-start; justify-content:space-between; gap:12px">
        <div>
          <h1 class="page-title">Users</h1>
          <p class="page-subtitle">Manage internal back-office users &amp; their access.</p>
        </div>
        <button class="btn-primary" @click="showForm = !showForm">
          {{ showForm ? 'Close' : '+ Add User' }}
        </button>
      </div>

      <form v-if="showForm" class="user-form" @submit.prevent="submit">
        <div class="user-form-grid">
          <div>
            <label class="login-label">Name</label>
            <input v-model="form.name" class="text-input" placeholder="Full name" required>
          </div>
          <div>
            <label class="login-label">Username</label>
            <input v-model="form.username" class="text-input" placeholder="e.g. johndoe" required>
          </div>
          <div>
            <label class="login-label">Email</label>
            <input v-model="form.email" type="email" class="text-input" placeholder="email@example.com" required>
          </div>
          <div>
            <label class="login-label">Password</label>
            <input v-model="form.password" type="password" class="text-input" placeholder="min. 6 characters" required>
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
          {{ saving ? 'Saving...' : 'Save User' }}
        </button>
      </form>

      <div class="table-card">
        <div class="table-scroll">
          <table class="data-table">
            <thead>
              <tr>
                <th style="width:60px">ID</th>
                <th>Name</th>
                <th>Username</th>
                <th>Email</th>
                <th>Role</th>
                <th>Status</th>
                <th>Created</th>
                <th class="text-center" style="width:120px">Action</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in users" :key="u.id">
                <td class="font-semibold">{{ u.id }}</td>
                <td>{{ u.name }}</td>
                <td class="font-semibold">{{ u.username }}</td>
                <td>{{ u.email }}</td>
                <td><span class="badge" :class="u.role === 'admin' ? 'badge-accent' : 'badge-muted'">{{ u.role }}</span></td>
                <td><span class="badge" :class="u.is_active ? 'badge-success' : 'badge-danger'">{{ u.is_active ? 'active' : 'inactive' }}</span></td>
                <td>{{ fmt(u.created_at) }}</td>
                <td class="text-center">
                  <button class="link-btn" @click="toggleActive(u)">
                    {{ u.is_active ? 'Deactivate' : 'Activate' }}
                  </button>
                </td>
              </tr>
              <tr v-if="!users.length">
                <td colspan="8" style="text-align:center; color:var(--text-muted); padding:28px">
                  No users yet
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
