import { getRouteApi, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import EventsFeed from '../components/EventsFeed'

const routeApi = getRouteApi('/namespaces/$ns')

interface Pod {
    name: string
    namespace: string
    phase: string
    host_node: string
    created_at: string
}

const fetchPods = async (ns: string): Promise<Pod[]> => {
    const res = await fetch('/api/v1/namespaces/' + ns + '/pods')
    if (!res.ok) throw new Error('Network error');
    return res.json()
}

function WorkloadsPage() {
  const { ns } = routeApi.useParams()
  const { data, isLoading, isError, error } = useQuery({
        queryKey: ['pods', ns],
        queryFn: () => fetchPods(ns),
    })

  return (
    <div>
      <Link to="/">← Namespaces</Link>
      {isLoading && <div>Loading...</div>}
      {isError && <div>Error: {error.message}</div>}
      <h2>{ns}</h2>
      <ul>
        {data?.map(pod => (
          <li key={pod.name}>
            <Link to="/namespaces/$ns/pods/$pn" params={{ns: ns, pn: pod.name}}>{pod.name}</Link> [{pod.phase}]
          </li>
        ))}
      </ul>
      <EventsFeed namespace={ns} />
    </div>
  )
}

export default WorkloadsPage