import {
  createRouter,
  createRoute,
  createRootRoute,
} from "@tanstack/react-router";
import App from "./App";
import NamespacesPage from "./pages/NamespacesPage";
import WorkloadsPage from "./pages/WorkloadsPage";
import DeploymentDetailsPage from "./pages/DeploymentDetailsPage";
import PodDetailsPage from "./pages/PodDetailsPage";

const rootRoute = createRootRoute({ component: App });

export const indexRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/",
  component: NamespacesPage,
});
export const workloadsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/namespaces/$ns",
  component: WorkloadsPage,
});
export const deploymentDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/namespaces/$ns/deployments/$name",
  component: DeploymentDetailsPage,
});
export const podDetailsRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: "/namespaces/$ns/pods/$name",
  component: PodDetailsPage,
});

export const routeTree = rootRoute.addChildren([
  indexRoute,
  workloadsRoute,
  deploymentDetailsRoute,
  podDetailsRoute,
]);

export const router = createRouter({ routeTree });
