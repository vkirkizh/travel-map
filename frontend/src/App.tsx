import {LandingPage} from "./pages/LandingPage";
import {PrivateAppPage} from "./pages/PrivateAppPage";
import {PublicMapPage} from "./pages/PublicMapPage";

const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080";

function App() {
  const pathname = window.location.pathname;

  if (pathname === "/" || pathname === "") {
    return <LandingPage />;
  }

  if (
    pathname === "/app" ||
    pathname === "/app/" ||
    pathname === "/app/settings" ||
    pathname === "/app/settings/"
  ) {
    return <PrivateAppPage apiBaseUrl={apiBaseUrl} pathname={pathname} />;
  }

  const username = pathname.replace(/^\/+|\/+$/g, "");

  return <PublicMapPage apiBaseUrl={apiBaseUrl} username={username} />;
}

export default App;
