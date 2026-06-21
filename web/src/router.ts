import { createRouter, createRoute, createRootRoute } from '@tanstack/react-router'
import App from './App'
import NamespaceList from './components/NamespaceList'
import WorkloadsPage from './pages/WorkloadsPage'

const rootRoute = createRootRoute({ component: App })

export const indexRoute = createRoute({ getParentRoute: () => rootRoute, path: '/', component: NamespaceList })
export const workloadsRoute = createRoute({ getParentRoute: () => rootRoute, path: '/namespaces/$ns', component: WorkloadsPage })

const routeTree = rootRoute.addChildren([indexRoute, workloadsRoute])

export const router = createRouter({ routeTree })