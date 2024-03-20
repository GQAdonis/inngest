'use client';

import { useEffect, useState } from 'react';
import { useAuth } from '@clerk/nextjs';
import {
  baseFetchSucceeded,
  baseInitialFetchFailed,
  baseInitialFetchLoading,
  type FetchResult,
} from '@inngest/components/types/fetch';

export function useRestAPIRequest<T>({
  url,
  method,
  pause = false,
}: {
  url: string | URL | null;
  method: string;
  pause: boolean;
}): FetchResult<T> {
  const { getToken } = useAuth();
  const [data, setData] = useState<any>();
  const [isLoading, setIsLoading] = useState<boolean>(false);
  const [error, setError] = useState<Error>();

  useEffect(() => {
    async function request() {
      if (!url || pause) return;
      const sessionToken = await getToken();
      if (!sessionToken) return; // TODO - Handle no auth

      setIsLoading(true);
      const response = await fetch(url, {
        method,
        headers: {
          Authorization: `Bearer ${sessionToken}`,
        },
      });
      if (!response.ok || response.status >= 400) {
        setData(null);
        setIsLoading(false);
        return setError(new Error(response.statusText));
      }
      const data = await response.json();

      setData(data);
      setIsLoading(false);
    }
    request();
  }, [getToken, url, method, pause]);

  if (isLoading) {
    return {
      ...baseInitialFetchLoading,
      isLoading: true,
    };
  }

  if (error) {
    return {
      ...baseInitialFetchFailed,
      error,
    };
  }

  if (!data) {
    // Should be unreachable.
    return {
      ...baseInitialFetchFailed,
      error: new Error('finished loading but missing data'),
    };
  }

  return {
    ...baseFetchSucceeded,
    data,
  };
}