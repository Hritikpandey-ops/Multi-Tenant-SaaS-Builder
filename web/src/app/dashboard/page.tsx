'use client';

import { useEffect, useState } from 'react';
import { apiClient } from '@/lib/api-client';
import { Tenant } from '@/types';

export default function DashboardPage() {
  const [loading, setLoading] = useState(true);
  const [usage, setUsage] = useState<any>(null);

  useEffect(() => {
    fetchUsage();
  }, []);

  const fetchUsage = async () => {
    try {
      const data = await apiClient.getTenantUsage();
      setUsage(data.data);
    } catch (error) {
      console.error('Failed to fetch usage:', error);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="mt-2 text-gray-600">Welcome to your dashboard</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Total Users</p>
              <p className="mt-2 text-3xl font-bold text-gray-900">
                {usage?.user_count || 0}
              </p>
            </div>
            <div className="w-12 h-12 bg-primary-100 rounded-lg flex items-center justify-center">
              <svg className="w-6 h-6 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 20 20">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z" />
              </svg>
            </div>
          </div>
          <div className="mt-4">
            <p className="text-xs text-gray-500">
              Limit: {usage?.user_limit === -1 ? 'Unlimited' : usage?.user_limit}
            </p>
            {usage?.user_limit !== -1 && (
              <div className="mt-2 bg-gray-200 rounded-full h-2">
                <div
                  className="bg-primary-600 h-2 rounded-full"
                  style={{
                    width: `${Math.min((usage?.user_count / usage?.user_limit) * 100, 100)}%`,
                  }}
                />
              </div>
            )}
          </div>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Current Plan</p>
              <p className="mt-2 text-3xl font-bold text-gray-900 capitalize">
                {usage?.plan || 'Free'}
              </p>
            </div>
            <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center">
              <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 20 20">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
          <p className="mt-4 text-xs text-gray-500">
            {usage?.plan === 'free' && 'Upgrade to Pro for more features'}
            {usage?.plan === 'pro' && 'You have access to all Pro features'}
            {usage?.plan === 'enterprise' && 'Enterprise plan active'}
          </p>
        </div>

        <div className="bg-white rounded-lg shadow p-6">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-sm font-medium text-gray-600">Tenant Status</p>
              <p className="mt-2 text-3xl font-bold text-green-600 capitalize">
                {usage?.status || 'Active'}
              </p>
            </div>
            <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center">
              <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 20 20">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
          </div>
          <p className="mt-4 text-xs text-gray-500">Your tenant is in good standing</p>
        </div>
      </div>

      {/* Quick Actions */}
      <div className="bg-white rounded-lg shadow p-6">
        <h2 className="text-lg font-semibold text-gray-900 mb-4">Quick Actions</h2>
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <a
            href="/dashboard/users"
            className="block p-4 border border-gray-200 rounded-lg hover:border-primary-500 hover:shadow-md transition-all"
          >
            <h3 className="font-medium text-gray-900">Invite User</h3>
            <p className="mt-1 text-sm text-gray-500">Add team members to your organization</p>
          </a>

          <a
            href="/dashboard/billing"
            className="block p-4 border border-gray-200 rounded-lg hover:border-primary-500 hover:shadow-md transition-all"
          >
            <h3 className="font-medium text-gray-900">Upgrade Plan</h3>
            <p className="mt-1 text-sm text-gray-500">Get more features and higher limits</p>
          </a>

          <a
            href="/dashboard/settings"
            className="block p-4 border border-gray-200 rounded-lg hover:border-primary-500 hover:shadow-md transition-all"
          >
            <h3 className="font-medium text-gray-900">Settings</h3>
            <p className="mt-1 text-sm text-gray-500">Manage your account settings</p>
          </a>
        </div>
      </div>
    </div>
  );
}
