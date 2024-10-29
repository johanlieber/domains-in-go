import {
  QueryClient,
  QueryClientProvider,
} from '@tanstack/solid-query'

const queryClient = new QueryClient();

export default function Layout(props) {
  return (
    <QueryClientProvider client={queryClient}>
      <main>
        {props.children}
      </main>
    </QueryClientProvider>
  )
}
