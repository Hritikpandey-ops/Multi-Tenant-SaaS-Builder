'use client';

import { ReactNode } from 'react';
import Link from 'next/link';
import { apiClient } from '@/lib/api-client';

export default function DashboardLayout({ children }: { children: ReactNode }) {
  const user = apiClient.getUser();
  const tenant = apiClient.getTenantFromStorage();

  const handleLogout = async () => {
    await apiClient.logout();
    window.location.href = '/';
  };

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Top Navigation */}
      <nav className="bg-white shadow-sm border-b border-gray-200">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16">
            <div className="flex items-center">
              <Link href="/dashboard" className="flex items-center">
                <span className="text-xl font-bold text-primary-600">
                  Multi-Tenant SaaS
                </span>
              </Link>
            </div>

            <div className="flex items-center space-x-4">
              <div className="text-sm text-gray-700">
                <span className="font-medium">{tenant?.name}</span>
                <span className="mx-2 text-gray-400">•</span>
                <span className="text-gray-500">{tenant?.plan}</span>
              </div>

              <div className="relative group">
                <button className="flex items-center space-x-2 text-sm text-gray-700 hover:text-gray-900">
                  <div className="w-8 h-8 bg-primary-600 rounded-full flex items-center justify-center text-white font-semibold">
                    {user?.first_name?.[0]}{user?.last_name?.[0]}
                  </div>
                  <span>{user?.first_name} {user?.last_name}</span>
                </button>

                {/* Dropdown Menu */}
                <div className="absolute right-0 mt-2 w-48 bg-white rounded-md shadow-lg py-1 opacity-0 invisible group-hover:opacity-100 group-hover:visible transition-all duration-200">
                  <Link
                    href="/dashboard/settings"
                    className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-100"
                  >
                    Settings
                  </Link>
                  <button
                    onClick={handleLogout}
                    className="block w-full text-left px-4 py-2 text-sm text-red-600 hover:bg-gray-100"
                  >
                    Sign out
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </nav>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="flex gap-8">
          {/* Sidebar */}
          <aside className="w-64 flex-shrink-0">
            <nav className="space-y-1">
              <Link
                href="/dashboard"
                className="block px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900"
              >
                Overview
              </Link>
              <Link
                href="/dashboard/users"
                className="block px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900"
              >
                Users
              </Link>
              <Link
                href="/dashboard/tenant"
                className="block px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900"
              >
                Tenant Settings
              </Link>
              <Link
                href="/dashboard/billing"
                className="block px-3 py-2 rounded-md text-sm font-medium text-gray-700 hover:bg-gray-100 hover:text-gray-900"
              >
                Billing
              </Link>
            </nav>
          </aside>

          {/* Main Content */}
          <main className="flex-1 min-w-0">
            {children}
          </main>
        </div>
      </div>
    </div>
  );
}
