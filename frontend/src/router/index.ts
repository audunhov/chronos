import { createRouter, createWebHistory } from 'vue-router'
import { auth } from '../services/auth'
import DashboardView from '../views/DashboardView.vue'
import AdminView from '../views/AdminView.vue'
import MemberListView from '../views/MemberListView.vue'
import TreasuryView from '../views/TreasuryView.vue'
import SurveysView from '../views/SurveysView.vue'
import ProfileView from '../views/ProfileView.vue'
import UserInspectorView from '../views/UserInspectorView.vue'

// Import components directly for other tabs to keep it simple
import OrgManager from '../components/OrgManager.vue'
import StatsDashboard from '../components/StatsDashboard.vue'
import FormManager from '../components/FormManager.vue'
import ReactionManager from '../components/ReactionManager.vue'
import OrganListManager from '../components/OrganListManager.vue'
import AuditTrail from '../components/AuditTrail.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: DashboardView,
    },
    {
      path: '/surveys',
      name: 'surveys',
      component: SurveysView,
    },
    {
      path: '/profile',
      name: 'profile',
      component: ProfileView,
    },
    {
      path: '/admin',
      component: AdminView,
      children: [
        { path: '', redirect: '/admin/members' },
        { path: 'members', component: MemberListView },
        { path: 'users/:id', component: UserInspectorView },
        { path: 'treasury', component: TreasuryView },
        { path: 'orgs', component: OrgManager },
        { path: 'stats', component: StatsDashboard },
        { path: 'forms', component: FormManager, props: { isAdmin: true } },
        { path: 'pipelines', component: ReactionManager },
        { path: 'organs', component: OrganListManager },
        { path: 'audit', component: AuditTrail },
      ]
    }
  ],
})

router.beforeEach((to) => {
  if (to.path.startsWith('/admin') && auth.user?.role !== 'admin') {
    return '/'
  }
})

export default router
