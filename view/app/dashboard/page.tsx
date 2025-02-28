"use client"
import DashboardLayout from '@/components/dashboard-layout'
import { useAppSelector } from '@/redux/hooks'
import React from 'react'

function page() {
  const user = useAppSelector(state => state.auth.user)

  return (
    <DashboardLayout>
      <h1>Dashboard</h1>
    </DashboardLayout>
  )
}

export default page