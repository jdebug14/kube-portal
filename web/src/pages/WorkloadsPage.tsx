import { getRouteApi, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import EventsFeed from '../components/EventsFeed'
import { apiFetch } from '../api/client'

const routeApi = getRouteApi('/namespaces/$ns')

interface Pod {
    name: string
    namespace: string
    phase: string
    host_node: string
    created_at: string
}

function WorkloadsPage() {
  const { ns } = routeApi.useParams()
  const url = `/api/v1/namespaces/${ns}/pods`
  const { data, isLoading, isError, error } = useQuery({
        queryKey: ['pods', ns],
        queryFn: () => apiFetch<Pod[]>(url, r => r.json()),
    })

  return (
    <div>
      <Link to="/">← Namespaces</Link>
      {isLoading && <div>Loading...</div>}
      {isError && <div>Error: {error.message}</div>}
      {data && (<div>
      <h2>{ns}</h2>
      <ul>
        {data.map(pod => (
          <li key={pod.name}>
            <Link to="/namespaces/$ns/pods/$pn" params={{ns: ns, pn: pod.name}}>{pod.name}</Link> [{pod.phase}]
          </li>
        ))}
      </ul>
      <EventsFeed namespace={ns} />
      </div>)}
    </div>
  )
}

export default WorkloadsPage