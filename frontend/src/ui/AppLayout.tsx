import { Outlet, Link, NavLink } from "react-router-dom";

export function AppLayout() {
  return (
    <div className="min-h-screen text-slate-100">
      <header className="border-b border-white/10 bg-slate-950/70 backdrop-blur">
        <div className="mx-auto flex max-w-6xl items-center justify-between px-6 py-5">
          <Link to="/" className="text-xl font-semibold tracking-wide text-amber-300">
            Tech Feed Hub
          </Link>
          <nav className="flex gap-6 text-sm text-slate-300">
            <NavLink to="/" end className={({ isActive }) => navLinkClassName(isActive)}>
              Articles
            </NavLink>
            <NavLink to="/sources" className={({ isActive }) => navLinkClassName(isActive)}>
              Sources
            </NavLink>
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-6 py-8">
        <Outlet />
      </main>
    </div>
  );
}

function navLinkClassName(isActive: boolean) {
  return isActive ? "text-white" : "transition hover:text-white";
}
