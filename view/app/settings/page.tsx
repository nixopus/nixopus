'use client';
import { useRouter } from 'next/navigation';

function Page() {
  const router = useRouter();
  router.push('/settings/general');
  return null;
}

export default Page;
