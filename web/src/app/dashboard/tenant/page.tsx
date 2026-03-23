'use client';

import Link from 'next/link';
import { apiClient } from '@/lib/api-client';

export default function TenantSettingsPage() {
  const user = apiClient.getUser();
  const tenant = apiClient.getTenantFromStorage();

  return (
    <div>
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Tenant Settings</h1>
        <p className="mt-2 text-gray-600">Manage your organization settings</p>
      </div>

      <div className="space-y-6">
        {/* Tenant Information */}
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Organization Information</h2>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Organization Name
              </label>
              <input
                type="text"
                defaultValue={tenant?.name}
                className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Organization Slug
              </label>
              <input
                type="text"
                defaultValue={tenant?.slug}
                disabled
                className="w-full px-3 py-2 border border-gray-300 rounded-lg bg-gray-50 text-gray-500"
              />
              <p className="mt-1 text-xs text-gray-500">Cannot be changed</p>
            </div>
          </div>

          <div className="mt-6">
            <button className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700">
              Save Changes
            </button>
          </div>
        </div>

        {/* Plan Information */}
        <div className="bg-white shadow rounded-lg p-6">
          <h2 className="text-lg font-semibold text-gray-900 mb-4">Current Plan</h2>

          <div className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
            <div>
              <p className="text-2xl font-bold text-gray-900 capitalize">{tenant?.plan}</p>
              <p className="text-sm text-gray-600 mt-1">
                {tenant?.plan === 'free' && '5 users, 100MB storage'}
                {tenant?.plan === 'pro' && '50 users, 10GB storage'}
                {tenant?.plan === 'enterprise' && 'Unlimited everything'}
              </p>
            </div>
            <Link
              href="/dashboard/billing"
              className="px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700"
            >
              Upgrade
            </Link>
          </div>
        </div>

        {/* Danger Zone */}
        <div className="bg-white shadow rounded-lg p-6 border-2 border-red-200">
          <h2 className="text-lg font-semibold text-red-600 mb-4">Danger Zone</h2>

          <div className="space-y-4">
            <div>
              <p className="text-sm text-gray-600 mb-2">
                Once you delete your organization, there is no going back. Please be certain.
              </p>
              <button className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700">
                Delete Organization
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
