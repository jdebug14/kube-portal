export async function apiFetch<T>(
  url: string,
  parse: (res: Response) => Promise<T>,
): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    const err = await res.json();
    throw new Error(err.error);
  }
  return parse(res);
}
