import Layout from '@/layouts/Layout'
import ky from "ky";
import { createMutation } from "@tanstack/solid-query";
import { Title } from "@solidjs/meta";
import { Match, Switch, For, createSignal } from "solid-js";

type PorkApiResponse = {
  domains: { tag: string; name: string; date: string; target: string; }[];
}

const ListedDomains = (props: { tag: string; name: string; date: string }) => {
  return (
    <section class='flex flex-row gap-x-10'>
      <span class='rounded px-3 bg-pink-300 font-extrabold text-red-500'>{props.tag}</span>
      <span>{props.name}</span>
      <span class='text-red-600'>{props.date}</span>
    </section>
  )
}

const ChangedDomains = (props: { tag: string; name: string; target: string }) => {
  return (
    <section class='flex flex-row gap-x-10'>
      <span class='rounded px-3 bg-pink-300 font-extrabold text-red-500'>{props.tag}</span>
      <span>{props.name}</span>
      <span class='text-red-600'>{props.target}</span>
    </section>
  )
}

const DefaultDomains = (domains: { status: string; name: string; expires_at: Date }[]) => {
  return (
    <For each={domains}>
      {(item) => <ListedDomains tag={item.status} name={item.name} date={item.expires_at.toLocaleString()} />}
    </For>
  )
}


const Domains = (props: { domains: { status: string; name: string; expires_at: Date }[] }) => {
  let [kind, setKind] = createSignal();
  const listing = "listing";
  const changed = "changed";
  const fetchListing = "fetch-listing";
  let { domains } = props;
  const porkApiSubmit = createMutation(() => ({
    mutationKey: ['pork-api'],
    mutationFn: () => ky.post<PorkApiResponse>('/domains', {
      json: {
        kind: kind()
      }
    }).json()
  }));
  // Setup Dynamic?
  return (
    <>
      <Title>Domains</Title>
      <section class="flex min-h-screen flex-col items-center gap-y-5 p-12">
        <form onsubmit={(e) => {
          e.preventDefault();
        }} class='w-2/4'>
          <section class='flex flex-col gap-y-3'>
            <label class='text-3xl font-extrabold text-rose-600' for="names">Names</label>
            <div class='flex flex-row gap-x-5'>
              <select onChange={() => {
                if (kind() !== listing) porkApiSubmit.mutate();
              }} onInput={(e) => setKind(e.currentTarget.value)} class='focus:ring-primary-600 focus:border-primary-600 dark:focus:ring-primary-500 dark:focus:border-primary-500 block w-full rounded-lg border border-gray-300 bg-gray-50 p-3 text-xl text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400' id="names" name="names">
                <option selected={true} onSelect={() => setKind(listing)} value={listing}>List Available Domains</option>
                <option value={changed}>Changed Domains</option>
                <option value={fetchListing}>Fetch New Domains</option>
              </select>
            </div>
          </section>
        </form>
        <span class='flex flex-col gap-y-5 w-1/2 min-w-fit rounded-lg bg-slate-200 min-h-20 h-auto px-4 py-3 text-3xl'>
          <Switch fallback={<DefaultDomains {...domains} />} >
            <Match when={porkApiSubmit.isSuccess && kind() === fetchListing}>
              <For each={porkApiSubmit.data.domains}>
                {(item) => <ListedDomains tag={item.tag} name={item.name} date={item.date} />}
              </For>
            </Match>
            <Match when={porkApiSubmit.isSuccess && kind() === changed}>
              <For each={porkApiSubmit.data.domains}>
                {(item) => <ChangedDomains tag={item.tag} name={item.name} target={item.target} />}
              </For>
            </Match>
            <Match when={porkApiSubmit.isPending}>
              <p class='p-2'>Loading...</p>
            </Match>
          </Switch>
        </span>
      </section>
    </>
  );
};

export default Domains;

Domains.layout = Layout;


