<template>
  <section class="card border border-primary-100 dark:border-primary-900/50">
    <div class="border-b border-gray-100 px-6 py-4 dark:border-dark-700">
      <div class="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
            {{ t("admin.settings.payment.billingMode.title") }}
          </h2>
          <p class="mt-1 max-w-3xl text-sm text-gray-500 dark:text-gray-400">
            {{ t("admin.settings.payment.billingMode.description") }}
          </p>
        </div>
        <span class="rounded-full bg-primary-50 px-3 py-1 text-xs font-medium text-primary-700 dark:bg-primary-900/30 dark:text-primary-300">
          {{ t("admin.settings.payment.billingMode.current", { currency: currentLabel }) }}
        </span>
      </div>
    </div>

    <div class="space-y-5 p-6">
      <div class="grid gap-5 lg:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t("admin.settings.payment.billingMode.settlementCurrency") }}
          </label>
          <select v-model="form.currency" class="form-input w-full max-w-sm">
            <option value="CNY">{{ t("admin.settings.payment.billingMode.cny") }}</option>
            <option value="USD">{{ t("admin.settings.payment.billingMode.usd") }}</option>
          </select>
          <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">{{ t("admin.settings.payment.billingMode.sourcePriceHint") }}</p>
        </div>
        <div>
          <label class="mb-2 block text-sm font-medium text-gray-700 dark:text-gray-300">
            {{ t("admin.settings.payment.billingMode.rate") }}
          </label>
          <div class="flex max-w-sm items-center gap-2">
            <input v-model.number="form.usd_to_cny_rate" type="number" min="0.0001" step="0.01" class="form-input w-full" />
            <span class="whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">CNY / USD</span>
          </div>
          <p class="mt-2 text-xs text-gray-500 dark:text-gray-400">{{ t("admin.settings.payment.billingMode.rateHint") }}</p>
        </div>
      </div>

      <div v-if="currencyChanged" class="rounded-lg border border-amber-200 bg-amber-50 p-4 dark:border-amber-900/60 dark:bg-amber-950/20">
        <div class="flex gap-3">
          <span class="mt-0.5 text-amber-600" aria-hidden="true">!</span>
          <div class="min-w-0 flex-1">
            <p class="font-medium text-amber-900 dark:text-amber-200">{{ t("admin.settings.payment.billingMode.switchTitle") }}</p>
            <p class="mt-1 text-sm leading-6 text-amber-800 dark:text-amber-300">{{ t("admin.settings.payment.billingMode.switchHint") }}</p>
            <div class="mt-3 grid gap-3 sm:grid-cols-3">
              <div class="rounded-md bg-white/70 p-3 dark:bg-dark-900/40">
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t("admin.settings.payment.billingMode.factor") }}</div>
                <div class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">×{{ conversionFactor.toFixed(6) }}</div>
              </div>
              <div class="rounded-md bg-white/70 p-3 dark:bg-dark-900/40 sm:col-span-2">
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ t("admin.settings.payment.billingMode.example") }}</div>
                <div class="mt-1 font-medium text-gray-900 dark:text-white">{{ conversionExample }}</div>
              </div>
            </div>
            <label class="mt-4 flex items-start gap-2 text-sm text-amber-900 dark:text-amber-200">
              <input v-model="backupConfirmed" type="checkbox" class="mt-1" />
              <span>{{ t("admin.settings.payment.billingMode.backupConfirm") }}</span>
            </label>
            <label class="mt-3 block text-sm text-amber-900 dark:text-amber-200">
              {{ t("admin.settings.payment.billingMode.confirmLabel", { phrase: confirmationPhrase }) }}
              <input v-model="confirmation" type="text" class="form-input mt-2 w-full max-w-md" :placeholder="confirmationPhrase" />
            </label>
          </div>
        </div>
      </div>

      <div class="flex flex-wrap items-center justify-between gap-3 border-t border-gray-100 pt-4 dark:border-dark-700">
        <p v-if="message" class="text-sm" :class="saveError ? 'text-red-600 dark:text-red-400' : 'text-emerald-600 dark:text-emerald-400'">{{ message }}</p>
        <span v-else class="text-xs text-gray-500 dark:text-gray-400">{{ t("admin.settings.payment.billingMode.historyHint") }}</span>
        <button type="button" class="btn btn-primary" :disabled="loading || saving || !canSave" @click="save">
          {{ saving ? t("common.saving") : t("common.save") }}
        </button>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from "vue";
import { useI18n } from "vue-i18n";
import { useAppStore } from "@/stores";
import { getBillingMode, updateBillingMode } from "@/api/admin/settings";
import { extractApiErrorMessage } from "@/utils/apiError";

type BillingCurrency = "USD" | "CNY";
const { t } = useI18n();
const appStore = useAppStore();
const loading = ref(true);
const saving = ref(false);
const saveError = ref(false);
const message = ref("");
const backupConfirmed = ref(false);
const confirmation = ref("");
const current = reactive<{ currency: BillingCurrency; rate: number }>({ currency: "CNY", rate: 7.2 });
const form = reactive<{ currency: BillingCurrency; usd_to_cny_rate: number }>({ currency: "CNY", usd_to_cny_rate: 7.2 });
const currencyChanged = computed(() => form.currency !== current.currency);
const currentLabel = computed(() => current.currency === "CNY" ? t("admin.settings.payment.billingMode.cny") : t("admin.settings.payment.billingMode.usd"));
const confirmationPhrase = computed(() => form.currency === "CNY" ? "SWITCH TO CNY" : "SWITCH TO USD");
const conversionFactor = computed(() => !currencyChanged.value ? 1 : form.currency === "CNY" ? Number(form.usd_to_cny_rate) : 1 / Number(current.rate || 1));
const conversionExample = computed(() => {
  if (!currencyChanged.value) return "100.00";
  const next = 100 * conversionFactor.value;
  return current.currency === "CNY" ? `¥100.00 → $${next.toFixed(2)}` : `$100.00 → ¥${next.toFixed(2)}`;
});
const canSave = computed(() => Number(form.usd_to_cny_rate) > 0 && (!currencyChanged.value || (backupConfirmed.value && confirmation.value.trim() === confirmationPhrase.value)));

async function load() {
  loading.value = true;
  try {
    const settings = await getBillingMode();
    current.currency = settings.currency;
    current.rate = settings.usd_to_cny_rate || 7.2;
    form.currency = current.currency;
    form.usd_to_cny_rate = current.rate;
  } catch (error) {
    saveError.value = true;
    message.value = extractApiErrorMessage(error);
  } finally {
    loading.value = false;
  }
}

async function save() {
  if (!canSave.value) return;
  const switched = currencyChanged.value;
  saving.value = true;
  saveError.value = false;
  message.value = "";
  try {
    const result = await updateBillingMode({ currency: form.currency, usd_to_cny_rate: Number(form.usd_to_cny_rate), convert_existing: switched, confirmation: switched ? confirmation.value.trim() : undefined });
    current.currency = result.currency;
    current.rate = result.usd_to_cny_rate || 7.2;
    form.currency = current.currency;
    form.usd_to_cny_rate = current.rate;
    backupConfirmed.value = false;
    confirmation.value = "";
    message.value = switched ? t("admin.settings.payment.billingMode.switchSuccess", { rows: result.rows_converted }) : t("admin.settings.payment.billingMode.saveSuccess");
    await appStore.fetchPublicSettings(true);
  } catch (error) {
    saveError.value = true;
    message.value = extractApiErrorMessage(error);
  } finally {
    saving.value = false;
  }
}

onMounted(load);
</script>
