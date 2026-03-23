# Multi-Tenant SaaS - Web Frontend

Next.js 15 frontend for the Multi-Tenant SaaS platform.

## Tech Stack

- **Framework**: Next.js 15 with App Router
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **HTTP Client**: Axios
- **Authentication**: JWT (localStorage)

## Getting Started

### Prerequisites

- Node.js 20+ installed
- Backend API running on http://localhost:3000

### Installation

```bash
cd web
npm install
```

### Development

```bash
npm run dev
```

Open [http://localhost:3001](http://localhost:3001) in your browser.

### Build for Production

```bash
npm run build
npm start
```

## Project Structure

```
web/
├── src/
│   ├── app/                 # Next.js App Router pages
│   │   ├── auth/            # Authentication pages
│   │   │   ├── login/
│   │   │   └── register/
│   │   ├── dashboard/       # Protected dashboard pages
│   │   │   ├── layout.tsx
│   │   │   ├── page.tsx
│   │   │   ├── users/
│   │   │   └── tenant/
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   └── globals.css
│   ├── components/         # Reusable components
│   ├── lib/                # Utility functions & API client
│   └── types/              # TypeScript types
├── public/                 # Static assets
├── package.json
├── tsconfig.json
├── tailwind.config.ts
└── next.config.js
```

## Features

### Authentication
- ✅ User registration with tenant creation
- ✅ Login with email/password
- ✅ JWT token management
- ✅ Auto token refresh
- ✅ Protected routes

### Dashboard
- ✅ Overview page with usage stats
- ✅ User management
- ✅ Tenant settings
- ✅ Plan information
- ✅ Quick actions

### UI Components
- Responsive design
- Dark mode ready
- Loading states
- Error handling
- Form validation

## API Integration

The frontend connects to the backend API:

```typescript
// API base URL (configured in .env.local)
NEXT_PUBLIC_API_URL=http://localhost:3000

// Example API call
import { apiClient } from '@/lib/api-client';

const data = await apiClient.register({
  email: 'user@example.com',
  password: 'password123',
  first_name: 'John',
  last_name: 'Doe',
  tenant_name: 'Acme Corp',
  tenant_slug: 'acme-corp'
});
```

## Environment Variables

Create `.env.local`:

```env
NEXT_PUBLIC_API_URL=http://localhost:3000
```

## Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server on port 3001 |
| `npm run build` | Build for production |
| `npm start` | Start production server |
| `npm run lint` | Run ESLint |

## Pages

| Route | Page | Protected |
|-------|------|-----------|
| `/` | Landing page | ❌ |
| `/auth/register` | Sign up | ❌ |
| `/auth/login` | Sign in | ❌ |
| `/dashboard` | Dashboard overview | ✅ |
| `/dashboard/users` | User management | ✅ |
| `/dashboard/tenant` | Tenant settings | ✅ |

## Future Enhancements

- [ ] Forgot password flow
- [ ] Email verification
- [ ] Profile settings page
- [ ] Billing management
- [ ] Analytics dashboard
- [ ] Real-time notifications
- [ ] Dark mode toggle
- [ ] Settings page

## Troubleshooting

### Port already in use
```bash
# Kill process on port 3001
lsof -ti:3001 | xargs kill -9
```

### Clean build
```bash
rm -rf .next node_modules
npm install
npm run dev
```

### API connection issues
- Ensure backend is running on http://localhost:3000
- Check `NEXT_PUBLIC_API_URL` in `.env.local`
