import Layout from "@/layouts/Layout";
import ky from "ky";
import { createMutation } from "@tanstack/solid-query";
import { Match, Switch, createSignal } from "solid-js";
import { Title } from "@solidjs/meta";

type Message = {
  message: string;
}

const DashBoard = (props: { url: string; }) => {
  const baseDomain = () => props.url;
  const [domain, setDomain] = createSignal(baseDomain());
  let subdomain: HTMLInputElement = null;
  let ttl: HTMLInputElement = null;
  let reqType: HTMLSelectElement = null;
  let prefix: HTMLInputElement = null;
  let host: HTMLInputElement = null;
  let description: HTMLTextAreaElement = null;
  const postFormData = createMutation(() => ({
    mutationKey: ['post-form'],
    mutationFn: () => ky.post<Message>('/data', {
      json: {
        subdomain: domain(), ttl: Number(ttl.value), kind: reqType.value, prefix: prefix.value, host: host.value, description: description.value
      }
    }).json()
  }));
  return (
    <>
      <Title>Dashboard</Title>
      <section class="bg-rose-300 min-h-screen dark:bg-gray-900">
        <div class="max-w-2xl px-4 py-8 mx-auto lg:py-16">
          <h2 class="mb-4 text-5xl font-bold text-slate-50">Register a Sub-Domain</h2>
          <Switch>
            <Match when={postFormData.isSuccess}>
              <span class="border-4 border-pink-500 bg-pink-500 w-full p-4 my-4 flex text-slate-50 text-xl">
                {postFormData.data.message}
              </span>
            </Match>
            <Match when={postFormData.isError}>
              <span class="border-4 border-red-500 bg-red-500 w-full p-4 my-4 flex text-slate-50 text-xl">
                {postFormData.error.message}
              </span>
            </Match>
            <Match when={postFormData.isPending}>
              <span class="border-4 border-yellow-500 bg-yellow-500 w-full p-4 my-4 flex text-slate-50 text-xl">
                loading...
              </span>
            </Match>
          </Switch>

          <form onSubmit={(e) => {
            e.preventDefault();
            console.log({ host: host.value, prefix: prefix.value, description: description.value })
            postFormData.mutate()
          }}>
            <div class="grid gap-4 mb-4 sm:grid-cols-2 sm:gap-6 sm:mb-5">
              <div class="sm:col-span-2">
                <label for="sub" class=" block mb-2 text-xl font-bold text-slate-50 dark:text-white">Subdomain</label>
                <input type="text" name="sub" id="name" class="bg-gray-50 border border-gray-300 text-gray-900 text-xl rounded-lg focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" placeholder="Type product name" disabled={true} ref={subdomain} value={domain()} />
              </div>
              <div class="w-full">
                <label for="prefix" class="block mb-2 text-lg font-bold text-slate-50 dark:text-white">Prefix</label>
                <input ref={prefix} onInput={(e) => {
                  const join = e.target.value ? "." : "";
                  setDomain(() => `${e.target.value}${join}${baseDomain()}`)
                }} autocomplete="off" type="text" name="prefix" id="prefix" class="bg-gray-50 border border-gray-300 text-gray-900 text-xl rounded-lg focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" placeholder="Product prefix" required={true} />
              </div>
              <div class="w-full">
                <label for="ttl" class="block mb-2 text-xl font-bold text-slate-50 dark:text-white">TTL (in mins)</label>
                <input ref={ttl} type="number" name="ttl" id="ttl" class="bg-gray-50 border border-gray-300 text-gray-900 text-xl rounded-lg focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" value={10} placeholder="10?" required={true} />
              </div>
              <div>
                <label for="domain-type" class="block mb-2 text-xl font-bold text-slate-50 dark:text-white">Type</label>
                <select ref={reqType} id="domain-type" name="options" class="bg-gray-50 border border-gray-300 text-gray-900 text-xl rounded-lg focus:ring-primary-500 focus:border-primary-500 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500">
                  <option selected={true} value="A">A Host</option>
                  <option value="CNAME">CNAME</option>
                  <option value="MX">MX</option>
                  <option value="TXT">TXT</option>
                </select>
              </div>
              <div>
                <label for="host-target-ip" class="block mb-2 text-xl font-bold text-slate-50 dark:text-white">Host Target (IP)</label>
                <input ref={host} autocomplete="off" type="text" pattern="^([0-9]{1,3}\.){3}[0-9]{1,3}$" name="host-target-ip" id="host-target-ip" class="bg-gray-50 border border-gray-300 text-gray-900 text-xl rounded-lg focus:ring-primary-600 focus:border-primary-600 block w-full p-2.5 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" placeholder="Ex. 12" required={true} />
              </div>
              <div class="sm:col-span-2">
                <label for="description" class="block mb-2 text-xl font-bold text-slate-50 dark:text-white">Description</label>
                <textarea ref={description} id="description" name="description" rows="8" class="block p-2.5 w-full text-xl text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-primary-500 focus:border-primary-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-primary-500 dark:focus:border-primary-500" placeholder="Write a product description here..."></textarea>
              </div>
            </div>
            <div class="flex items-center space-x-8">
              <button type="submit" class="text-red-600 inline-flex items-center hover:text-white border border-red-600 hover:bg-red-600 focus:ring-4 focus:outline-none focus:ring-red-300 font-bold rounded-lg text-2xl px-5 py-2.5 text-center dark:border-red-500 dark:text-red-500 dark:hover:text-white dark:hover:bg-red-600 dark:focus:ring-red-900 bg-slate-100" disabled={postFormData.isPending}>
                <svg class="pr-1" fill="currentColor" stroke-width="0" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512" style="overflow: visible;" height="1em" width="1em"><path d="M256 48a208 208 0 1 1 0 416 208 208 0 1 1 0-416zm0 464a256 256 0 1 0 0-512 256 256 0 1 0 0 512zM135.1 217.4c-4.5 4.2-7.1 10.1-7.1 16.3 0 12.3 10 22.3 22.3 22.3H208v96c0 17.7 14.3 32 32 32h32c17.7 0 32-14.3 32-32v-96h57.7c12.3 0 22.3-10 22.3-22.3 0-6.2-2.6-12.1-7.1-16.3l-107.1-99.9c-3.8-3.5-8.7-5.5-13.8-5.5s-10.1 2-13.8 5.5l-107.1 99.9z"></path></svg>
                Update
              </button>
            </div>
          </form>
        </div>
      </section>
    </>
  )
}

export default DashBoard;

DashBoard.layout = Layout;
