import Layout from '@/layouts/Layout'
import ky from "ky";
import { createMutation } from "@tanstack/solid-query";
import { Title } from "@solidjs/meta";
import { Match, Switch, For } from "solid-js";

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


const Domains = () => {
  let kind: HTMLSelectElement;
  const listing = "listing";
  const changed = "changed";
  const porkApiSubmit = createMutation(() => ({
    mutationKey: ['pork-api'],
    mutationFn: () => ky.post<PorkApiResponse>('/domains', {
      json: {
        kind: kind.value
      }
    }).json()
  }));
  return (
    <>
      <Title>Domains</Title>
      <section class="flex min-h-screen flex-col items-center gap-y-5 p-12">
        <form onsubmit={(e) => {
          e.preventDefault();
          porkApiSubmit.mutate();
        }} class='w-2/4'>
          <section class='flex flex-col gap-y-3'>
            <label class='text-3xl font-extrabold text-rose-600' for="names">Names</label>
            <div class='flex flex-row gap-x-5'>
              <select ref={kind} class='focus:ring-primary-600 focus:border-primary-600 dark:focus:ring-primary-500 dark:focus:border-primary-500 block w-full rounded-lg border border-gray-300 bg-gray-50 p-3 text-xl text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400' id="names" name="names">
                <option value={listing}>List Available Domains</option>
                <option value={changed}>Changed Domains</option>
              </select>
              <button class='border-2 rounded-lg p-4 bg-rose-600 text-white text-lg font-extrabold'>Submit</button>
            </div>
          </section>
        </form>
        <span class='flex flex-col gap-y-5 w-1/2 min-w-fit rounded-lg bg-slate-200 min-h-20 h-auto px-4 py-3 text-3xl'>
          <Switch>
            <Match when={porkApiSubmit.isSuccess && kind.value === listing}>
              <For each={porkApiSubmit.data.domains}>
                {(item) => <ListedDomains tag={item.tag} name={item.name} date={item.date} />}
              </For>
            </Match>
            <Match when={porkApiSubmit.isSuccess && kind.value === changed}>
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


