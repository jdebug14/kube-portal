import { createRouter, createRoute, createRootRoute } from '@tanstack/react-router'
import App from './App'
import NamespaceList from './components/NamespaceList'
import WorkloadsPage from './pages/WorkloadsPage'
import PodDetailsPage from './pages/PodDetailsPage'

const rootRoute = createRootRoute({ component: App })

export const indexRoute = createRoute({ getParentRoute: () => rootRoute, path: '/', component: NamespaceList })
export const workloadsRoute = createRoute({ getParentRoute: () => rootRoute, path: '/namespaces/$ns', component: WorkloadsPage })
export const podDetailsRoute = createRoute({ getParentRoute: () => rootRoute, path: '/namespaces/$ns/pods/$pn', component: PodDetailsPage })

const routeTree = rootRoute.addChildren([indexRoute, workloadsRoute, podDetailsRoute])

export const router = createRouter({ routeTree })