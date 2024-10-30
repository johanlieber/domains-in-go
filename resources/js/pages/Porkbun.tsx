import Layout from '@/layouts/Layout'
import ky from "ky";
import { createMutation } from "@tanstack/solid-query";
import { Title } from "@solidjs/meta";

const Domains = () => {
  const porkApiSubmit = createMutation(() => ({
    mutationKey: ['pork-api'],
    mutationFn: () => ky.post('/domains', {
      json: {

      }
    }).json()
  }));
  return (
    <>
      <Title>Domains</Title>
      <section class="flex min-h-screen flex-col items-center gap-y-5 p-12">
        <form onsubmit={(e) => e.preventDefault()} class='w-2/4'>
          <section class='flex flex-col gap-y-3'>
            <label class='text-3xl font-extrabold text-rose-600' for="names">Names</label>
            <div class='flex flex-row gap-x-5'>
              <select class='focus:ring-primary-600 focus:border-primary-600 dark:focus:ring-primary-500 dark:focus:border-primary-500 block w-full rounded-lg border border-gray-300 bg-gray-50 p-3 text-xl text-gray-900 dark:border-gray-600 dark:bg-gray-700 dark:text-white dark:placeholder-gray-400' id="names" name="names">
                <option value="listing">List Available Domains</option>
                <option value="changed">Changed Domains</option>
              </select>
              <button class='border-2 rounded-lg p-4 bg-rose-600 text-white text-lg font-extrabold'>Submit</button>
            </div>
          </section>
        </form>
        <span class='w-2/4 rounded-lg bg-slate-200 h-96 px-4 py-3 font-bold text-3xl'>
          info
        </span>
      </section>
    </>
  );
};

export default Domains;

Domains.layout = Layout;


