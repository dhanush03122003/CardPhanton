import { Link, useLocation } from "react-router-dom";

function Navbar({ user, onLogout, darkMode, onToggleTheme }) {
  const location = useLocation();
  const path = location.pathname;

  return (
    <nav className='bg-white border-b border-gray-200 dark:bg-slate-950 dark:border-slate-800'>
      <div className='max-w-7xl mx-auto px-4 sm:px-6 lg:px-8'>
        <div className='flex justify-between h-16'>
          <div className='flex items-center'>
            <Link to='/' className='flex items-center space-x-2 group'>
              <div className='w-8 h-8 bg-slate-900 rounded-lg flex items-center justify-center group-hover:bg-slate-800 transition-colors shadow-sm'>
                <svg
                  className='w-5 h-5 text-white'
                  fill='none'
                  stroke='currentColor'
                  viewBox='0 0 24 24'
                >
                  <path
                    strokeLinecap='round'
                    strokeLinejoin='round'
                    strokeWidth={2}
                    d='M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z'
                  />
                </svg>
              </div>
              <span className='text-xl font-bold text-gray-900 tracking-tight dark:text-white'>
                WebAuthn
              </span>
            </Link>
          </div>
          <div className='flex items-center space-x-3'>
            <button
              type='button'
              onClick={onToggleTheme}
              className='relative flex h-9 w-[68px] items-center rounded-full border border-gray-200 bg-gray-100 p-1 transition-colors hover:border-gray-300 dark:border-slate-700 dark:bg-slate-800 dark:hover:border-slate-600'
              aria-label={
                darkMode ? "Switch to light mode" : "Switch to dark mode"
              }
              title={darkMode ? "Switch to light mode" : "Switch to dark mode"}
              aria-pressed={darkMode}
            >
              <span
                className='absolute inset-y-1 left-1 flex w-7 items-center justify-center text-amber-500 transition-opacity dark:text-slate-400'
                aria-hidden='true'
              >
                <svg
                  className='h-4 w-4'
                  viewBox='0 0 24 24'
                  fill='none'
                  stroke='currentColor'
                  strokeWidth='2'
                >
                  <circle cx='12' cy='12' r='4' />
                  <path d='M12 2v2m0 16v2M4.93 4.93l1.41 1.41m11.32 11.32l1.41 1.41M2 12h2m16 0h2M4.93 19.07l1.41-1.41M17.66 6.34l1.41-1.41' />
                </svg>
              </span>
              <span
                className='absolute inset-y-1 right-1 flex w-7 items-center justify-center text-slate-500 dark:text-slate-300'
                aria-hidden='true'
              >
                <svg
                  className='h-4 w-4'
                  viewBox='0 0 24 24'
                  fill='none'
                  stroke='currentColor'
                  strokeWidth='2'
                >
                  <path d='M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79Z' />
                </svg>
              </span>
              <span
                className={`relative z-10 h-7 w-7 rounded-full bg-white shadow-sm ring-1 ring-black/5 transition-transform duration-200 dark:bg-slate-950 dark:ring-white/10 ${darkMode ? "translate-x-7" : "translate-x-0"}`}
              />
            </button>
            {user ? (
              <>
                <span className='text-sm font-medium text-gray-500 dark:text-slate-400'>
                  Welcome,{" "}
                  <span className='text-gray-900 dark:text-white'>
                    {user.username}
                  </span>
                </span>
                <div className='w-px h-4 bg-gray-200 mx-2 dark:bg-slate-700'></div>
                <button
                  onClick={onLogout}
                  className='px-3 py-2 text-sm font-medium text-gray-600 hover:text-slate-900 hover:bg-gray-50 rounded-lg transition-colors dark:text-slate-300 dark:hover:bg-slate-800 dark:hover:text-white'
                >
                  Log out
                </button>
              </>
            ) : (
              <div className='flex items-center space-x-2 bg-gray-50/50 p-1 rounded-lg border border-gray-100 dark:bg-slate-900 dark:border-slate-800'>
                <Link
                  to='/login'
                  className={`px-4 py-1.5 text-sm font-medium rounded-md transition-all duration-200 ${
                    path === "/login"
                      ? "bg-white text-slate-900 shadow-sm ring-1 ring-gray-200/50 dark:bg-slate-800 dark:text-white dark:ring-slate-700"
                      : "text-gray-500 hover:text-slate-900 hover:bg-gray-100/50 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-white"
                  }`}
                >
                  Log in
                </Link>
                <Link
                  to='/register'
                  className={`px-4 py-1.5 text-sm font-medium rounded-md transition-all duration-200 ${
                    path === "/register"
                      ? "bg-white text-slate-900 shadow-sm ring-1 ring-gray-200/50 dark:bg-slate-800 dark:text-white dark:ring-slate-700"
                      : "text-gray-500 hover:text-slate-900 hover:bg-gray-100/50 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-white"
                  }`}
                >
                  Register
                </Link>
              </div>
            )}
          </div>
        </div>
      </div>
    </nav>
  );
}

export default Navbar;
