import { Outlet } from "@tanstack/react-router"

function App() {
  return (
    <div>
      <h1>KubePortal</h1>
      <Outlet />
    </div>
  )
}

export default App