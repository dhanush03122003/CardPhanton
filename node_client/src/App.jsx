import { useState, useEffect } from "react";
import { Routes, Route, Navigate, useNavigate } from "react-router-dom";
import Navbar from "./components/Navbar";
import Register from "./pages/Register";
import Login from "./pages/Login";
import Dashboard from "./pages/Dashboard";

function App() {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const navigate = useNavigate();

  useEffect(() => {
    // Verify authentication status via cookies on app load
    verifyAuthentication();
  }, []);

  const verifyAuthentication = async () => {
    try {
      const response = await fetch("/api/auth/verify-token", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        credentials: "include", // Include HTTP-only cookies
      });

      if (response.ok) {
        const data = await response.json();
        // Extract user object from nested response structure
        setUser(data.user);
      }
    } catch (error) {
      console.error("Error verifying authentication:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleLogin = (user) => {
    setUser(user);
    navigate("/dashboard");
  };

  const handleLogout = async () => {
    try {
      // Call logout endpoint to clear server-side cookies
      await fetch("/api/auth/logout", {
        method: "POST",
        credentials: "include",
      });
    } catch (error) {
      console.error("Error during logout:", error);
    } finally {
      setUser(null);
      navigate("/login");
    }
  };

  if (loading) {
    return (
      <div className='min-h-screen flex items-center justify-center'>
        <div className='animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600'></div>
      </div>
    );
  }

  return (
    <div className='min-h-screen'>
      <Navbar user={user} onLogout={handleLogout} />
      <Routes>
        <Route
          path='/register'
          element={user ? <Navigate to='/dashboard' /> : <Register />}
        />
        <Route
          path='/login'
          element={
            user ? (
              <Navigate to='/dashboard' />
            ) : (
              <Login onLoginSuccess={handleLogin} />
            )
          }
        />
        <Route
          path='/dashboard'
          element={user ? <Dashboard user={user} /> : <Navigate to='/login' />}
        />
        <Route
          path='/'
          element={<Navigate to={user ? "/dashboard" : "/login"} />}
        />
      </Routes>
    </div>
  );
}

export default App;
