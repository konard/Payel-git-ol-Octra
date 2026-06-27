import { redirect } from 'next/navigation';
import { ROUTES } from '../../config/routes';

export default function ApiRedirect() {
  redirect(`${ROUTES.PROFILE_API}/public`);
}
