import axios from "@/utils/axios";
import {
  requireApiArrayField,
  requireApiBooleanField,
  requireApiObject,
  requireApiNumberField,
  requireApiStringField,
  unwrapApiPayload,
} from "@/utils/apiResponse";

export interface PaymentProviderInstallmentsSettings {
  provider: string;
  enabled: boolean;
  payment_method_types?: string[];
  countries?: string[];
  currencies?: string[];
  min_amount?: number;
  max_amount?: number;
  notes?: string;
}

export interface PaymentProviderInstallmentsUpdateRequest {
  enabled: boolean;
  payment_method_types: string[];
  countries: string[];
  currencies: string[];
  min_amount: number;
  max_amount: number;
  notes: string;
}

const endpoint = "/api/admin/settings/payment-installments";

const readPaymentProviderInstallmentsSettings = (
  response: unknown,
  path: string,
): PaymentProviderInstallmentsSettings => {
  const payload = requireApiObject(unwrapApiPayload(response, path), path);
  const settings: PaymentProviderInstallmentsSettings = {
    provider: requireApiStringField(payload, "provider", path),
    enabled: requireApiBooleanField(payload, "enabled", path),
  };

  if (Object.prototype.hasOwnProperty.call(payload, "payment_method_types")) {
    settings.payment_method_types = requireApiArrayField<string>(
      payload,
      "payment_method_types",
      path,
    );
  }
  if (Object.prototype.hasOwnProperty.call(payload, "countries")) {
    settings.countries = requireApiArrayField<string>(
      payload,
      "countries",
      path,
    );
  }
  if (Object.prototype.hasOwnProperty.call(payload, "currencies")) {
    settings.currencies = requireApiArrayField<string>(
      payload,
      "currencies",
      path,
    );
  }
  if (Object.prototype.hasOwnProperty.call(payload, "min_amount")) {
    settings.min_amount = requireApiNumberField(payload, "min_amount", path);
  }
  if (Object.prototype.hasOwnProperty.call(payload, "max_amount")) {
    settings.max_amount = requireApiNumberField(payload, "max_amount", path);
  }
  if (Object.prototype.hasOwnProperty.call(payload, "notes")) {
    settings.notes = requireApiStringField(payload, "notes", path);
  }

  return settings;
};

export const paymentInstallmentsApi = {
  async get(provider: string): Promise<PaymentProviderInstallmentsSettings> {
    const path = `${endpoint}/${provider}`;
    const response = await axios.get(path);
    return readPaymentProviderInstallmentsSettings(response, path);
  },

  async update(
    provider: string,
    payload: PaymentProviderInstallmentsUpdateRequest,
  ): Promise<PaymentProviderInstallmentsSettings> {
    const path = `${endpoint}/${provider}`;
    const response = await axios.put(path, payload);
    return readPaymentProviderInstallmentsSettings(response, path);
  },

  async remove(provider: string): Promise<void> {
    await axios.delete(`${endpoint}/${provider}`);
  },
};

export default paymentInstallmentsApi;
