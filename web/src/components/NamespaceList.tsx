import { useQuery } from '@tanstack/react-query'

interface Namespace {
    name: string
    status: string
    created_at: string
}

const fetchNamespaces = async (): Promise<Namespace[]> => {
    const res = await fetch('/api/v1/namespaces')
    if (!res.ok) throw new Error('Network error');
    return res.json()
}

function NamespaceList() {
    const { data, isLoading, isError, error } = useQuery({
        queryKey: ['namespaces'],
        queryFn: fetchNamespaces,
    })
    
    if (isLoading) return <div>Loading...</div>
    if (isError) return <div>Error: {error.message}</div>

    return (
        <ul>
            {data?.map(ns => (
                <li key={ns.name}>{ns.name}</li>
            ))}
        </ul>
    )
}

export default NamespaceList