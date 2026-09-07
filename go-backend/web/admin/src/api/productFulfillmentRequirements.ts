import axios from '@/utils/axios'
import {
  requireApiArrayField,
  requireApiNumberField,
  requireApiObject,
  unwrapApiPayload,
} from '@/utils/apiResponse'
import type {
  ProductFulfillmentRequirementListResult,
  ProductFulfillmentRequirementRule,
  ProductFulfillmentRequirementUpsertInput,
} from '@/modules/product/productFulfillmentRequirementTypes'

const productPath = (productID: number | string): string => (
  `/api/admin/products/${encodeURIComponent(String(productID))}/fulfillment-requirements`
)

const readObjectPayload = (response: unknown, path: string) => (
  requireApiObject(unwrapApiPayload(response, path), path)
)

const readRule = (value: unknown, path: string): ProductFulfillmentRequirementRule => (
  requireApiObject(value, path, 'rule') as ProductFulfillmentRequirementRule
)

export const productFulfillmentRequirementApi = {
  async list(productID: number | string): Promise<ProductFulfillmentRequirementListResult> {
    const path = productPath(productID)
    const payload = readObjectPayload(await axios.get(path), path)
    return {
      product_id: requireApiNumberField(payload, 'product_id', path),
      rules: requireApiArrayField<ProductFulfillmentRequirementRule>(payload, 'rules', path),
    }
  },

  async upsert(
    productID: number | string,
    input: ProductFulfillmentRequirementUpsertInput,
  ): Promise<ProductFulfillmentRequirementRule> {
    const path = `${productPath(productID)}/spoke-tension-qc`
    return readRule(await axios.put(path, input), path)
  },
}

export default productFulfillmentRequirementApi
