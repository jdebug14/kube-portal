import { getRouteApi } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import { Link } from '@tanstack/react-router'

const routeApi = getRouteApi('/namespaces/$ns')

interface Pod {
    name: string
    namespace: string
    phase: number
    host_node: number
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

  if (isLoading) return <div>Loading...</div>
  if (isError) return <div>Error: {error.message}</div>
  return (
    <div>
      <Link to="/">← Namespaces</Link>
      <h2>{ns}</h2>
      <ul>
        {data?.map(pod => (
          <li key={pod.name}>{pod.name} [{pod.phase}]</li>
        ))}
      </ul>
    </div>
  )
}

export default WorkloadsPage