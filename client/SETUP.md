# Frontend Setup Instructions

## Install Required Dependencies

You need to install a few Radix UI components for the auth page to work:

```bash
cd client
npm install @radix-ui/react-tabs @radix-ui/react-label
```

## Start the Development Server

```bash
npm run dev
```

The frontend will be available at `http://localhost:5173`

## Routes Available

- `/` - Landing page
- `/auth` - Login/Signup page (NEW!)
- `/dashboard` - Dashboard (requires authentication)

## Features

### Authentication Page (`/auth`)

The auth page includes:

✅ **Tabbed Interface** - Switch between Login and Signup
✅ **Form Validation** - Client-side validation before submission
✅ **Error Handling** - Clear error messages
✅ **Loading States** - Spinners during API calls
✅ **Responsive Design** - Works on mobile and desktop
✅ **Beautiful UI** - Modern gradient design with animations

### Login Form

Fields:

- Email or Username
- Password

### Signup Form

Fields:

- First Name
- Last Name
- Username
- Email
- Password
- Confirm Password

### API Integration

The auth page connects to your backend at `http://localhost:3000`:

- `POST /api/user/login` - Login endpoint
- `POST /api/user/signup` - Signup endpoint
- `POST /api/user/logout` - Logout endpoint

### Authentication Context

The `AuthContext` provides:

- `user` - Current user object
- `isAuthenticated` - Boolean for auth status
- `isLoading` - Loading state
- `login(email, password)` - Login function
- `signup(...)` - Signup function
- `logout()` - Logout function

### Using Auth in Components

```tsx
import { useAuth } from "~/contexts/AuthContext";

function MyComponent() {
  const { user, isAuthenticated, logout } = useAuth();

  if (!isAuthenticated) {
    return <div>Please login</div>;
  }

  return (
    <div>
      <h1>Welcome {user.firstName}!</h1>
      <button onClick={logout}>Logout</button>
    </div>
  );
}
```

## Next Steps

1. Install dependencies
2. Start the dev server
3. Navigate to `/auth`
4. Create an account or login
5. Build your dashboard!

## Environment Variables

Create a `.env` file in the `client` directory if you need to change the API URL:

```bash
VITE_API_URL=http://localhost:3000
```

Then update `AuthContext.tsx`:

```tsx
const API_URL = import.meta.env.VITE_API_URL || "http://localhost:3000";
```
