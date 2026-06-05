
  import { createRoot } from "react-dom/client";
  import App from "./app/App.tsx";
  import { DesktopTitleBar } from "./desktop/DesktopTitleBar";
  import "./styles/index.css";

  // The desktop title bar renders only inside the Electron shell (it returns null
  // in a browser), so the web build is unaffected. Mounting it here — above App —
  // keeps the frameless-window controls present across every App view, including
  // the landing and payment-success screens that App returns early.
  createRoot(document.getElementById("root")!).render(
    <div className="flex h-screen flex-col overflow-hidden">
      <DesktopTitleBar />
      <div className="min-h-0 flex-1">
        <App />
      </div>
    </div>
  );
