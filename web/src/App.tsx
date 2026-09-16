import React from 'react';
import { Routes, Route } from 'react-router-dom';
import { UserMiniApp } from './views/UserMiniApp';
import { AuthorStudio } from './views/AuthorStudio';
import { AdminDashboard } from './views/AdminDashboard';
import { AboutView } from './views/AboutView';
import { OnboardingWizard } from './views/OnboardingWizard';
import { ProtectedRoute } from './components/ProtectedRoute';
import { SetupGuard } from './components/SetupGuard';
import { Header } from './components/common/Header';
import { Footer } from './components/common/Footer';

export const App: React.FC = () => {
  return (
    <div className="min-h-screen bg-[#060911] text-slate-100 flex flex-col relative overflow-hidden">
      {/* Ambient Liquid Lighting Glow Orbs */}
      <div className="ambient-orb w-[600px] h-[600px] bg-cyan-600/15 -top-40 -left-40" />
      <div className="ambient-orb w-[500px] h-[500px] bg-blue-600/15 top-1/3 -right-40" style={{ animationDelay: '-6s' }} />
      <div className="ambient-orb w-[550px] h-[550px] bg-indigo-600/15 -bottom-40 left-1/3" style={{ animationDelay: '-12s' }} />

      {/* Universal Modular Header */}
      <Header />

      {/* Main Viewport */}
      <main className="flex-1 max-w-7xl w-full mx-auto p-3 sm:p-6 md:p-8 relative z-10">
        <SetupGuard>
          <Routes>
            {/* User Mini App & Media Catalog */}
            <Route path="/" element={<UserMiniApp />} />
            <Route path="/catalog" element={<UserMiniApp />} />
            <Route path="/vip" element={<UserMiniApp />} />
            <Route path="/watch/:slug" element={<UserMiniApp />} />

            {/* Project About Page */}
            <Route path="/about" element={<AboutView />} />

            {/* Setup / Onboarding Wizard Flow */}
            <Route path="/setup" element={<OnboardingWizard />} />
            <Route path="/onboarding" element={<OnboardingWizard />} />

            {/* Author Studio Pages */}
            <Route
              path="/studio"
              element={
                <ProtectedRoute allowedRoles={['author', 'admin', 'owner']}>
                  <AuthorStudio />
                </ProtectedRoute>
              }
            />
            <Route
              path="/studio/:tab"
              element={
                <ProtectedRoute allowedRoles={['author', 'admin', 'owner']}>
                  <AuthorStudio />
                </ProtectedRoute>
              }
            />

            {/* Admin Dashboard & Operation Pages */}
            <Route
              path="/admin"
              element={
                <ProtectedRoute allowedRoles={['admin', 'owner']}>
                  <AdminDashboard />
                </ProtectedRoute>
              }
            />
            <Route
              path="/admin/:tab"
              element={
                <ProtectedRoute allowedRoles={['admin', 'owner']}>
                  <AdminDashboard />
                </ProtectedRoute>
              }
            />
          </Routes>
        </SetupGuard>
      </main>

      {/* Universal Modular Footer */}
      <Footer />
    </div>
  );
};
