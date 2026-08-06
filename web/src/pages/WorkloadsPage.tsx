import { getRouteApi, Link } from "@tanstack/react-router";

import EventsFeed from "../components/EventsFeed";
import Pods from "../components/Pods";
import Deployments from "../components/Deployments";

const routeApi = getRouteApi("/namespaces/$ns");

export default function WorkloadsPage() {
  const { ns } = routeApi.useParams();
  return (
    <>
      <Link to="/">← namespaces</Link>
      <h2>Namespace: {ns}</h2>
      <Deployments namespace={ns} />
      <Pods namespace={ns} />
      <EventsFeed namespace={ns} />
    </>
  );
}
