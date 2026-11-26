<template>
  <AdminFormSection id="wheelset-product-specification" title="WHEELSET 产品级规格" description="分别记录前轮和后轮的制造商产品事实；没有可靠资料的可选字段留空。">
    <div class="grid gap-x-8 gap-y-6 lg:grid-cols-2">
      <section class="min-w-0 space-y-4" aria-labelledby="wheelset-front-wheel-heading">
        <div class="border-b pb-2">
          <h3 id="wheelset-front-wheel-heading" class="text-sm font-semibold">前轮</h3>
        </div>
        <fieldset :disabled="locked" class="grid min-w-0 gap-4 sm:grid-cols-2">
          <AdminFormField label="轮径" required description="填写制造商标注的市场轮径，例如 700C 或 29 in。">
            <Input v-model="specification.front_wheel.wheel_size" maxlength="32" placeholder="700C" />
          </AdminFormField>
          <AdminFormField label="支持孔数" required description="多个值用逗号分隔，例如 24, 28。">
            <Input v-model="specification.front_wheel.spoke_hole_counts" placeholder="24, 28" />
          </AdminFormField>
          <AdminFormField label="支持编法" required description="多个值用逗号分隔，例如 2x, 3x, radial。">
            <Input v-model="specification.front_wheel.lacing_patterns" placeholder="2x, 3x" />
          </AdminFormField>
          <AdminFormField label="BSD（mm）"><Input v-model="specification.front_wheel.bead_seat_diameter_mm" type="number" min="0.01" step="any" /></AdminFormField>
          <AdminFormField label="材质"><Input v-model="specification.front_wheel.material" maxlength="120" /></AdminFormField>
          <AdminFormField label="制动接口"><Input v-model="specification.front_wheel.brake_interface" maxlength="80" placeholder="Center Lock / 6-bolt / rim brake" /></AdminFormField>
          <AdminFormField label="胎圈结构"><Input v-model="specification.front_wheel.rim_profile" maxlength="80" placeholder="Hooked / Hookless" /></AdminFormField>
          <AdminFormField label="轮胎系统"><Input v-model="specification.front_wheel.tire_system" maxlength="80" placeholder="Clincher / Tubeless / Tubular" /></AdminFormField>
          <AdminFormField label="框高（mm）"><Input v-model="specification.front_wheel.rim_depth_mm" type="number" min="0.01" step="any" /></AdminFormField>
          <AdminFormField label="内宽（mm）"><Input v-model="specification.front_wheel.rim_internal_width_mm" type="number" min="0.01" step="any" /></AdminFormField>
          <AdminFormField label="外宽（mm）"><Input v-model="specification.front_wheel.rim_external_width_mm" type="number" min="0.01" step="any" /></AdminFormField>
          <AdminFormField label="花鼓接口"><Input v-model="specification.front_wheel.hub_interface" maxlength="120" /></AdminFormField>
          <AdminFormField label="轴规格"><Input v-model="specification.front_wheel.axle_standard" maxlength="120" /></AdminFormField>
          <AdminFormField label="辐条类型"><Input v-model="specification.front_wheel.spoke_type" maxlength="120" /></AdminFormField>
          <AdminFormField label="前轮重量（g）"><Input v-model="specification.front_wheel.weight_g" type="number" min="0.01" step="any" /></AdminFormField>
          <AdminFormField label="最大胎压（bar）"><Input v-model="specification.front_wheel.max_tire_pressure_bar" type="number" min="0.01" step="any" /></AdminFormField>
        </fieldset>
      </section>

      <section class="min-w-0 space-y-4" aria-labelledby="wheelset-rear-wheel-heading">
        <div class="border-b pb-2">
          <h3 id="wheelset-rear-wheel-heading" class="text-sm font-semibold">后轮</h3>
        </div>
        <fieldset :disabled="locked" class="grid min-w-0 gap-4 sm:grid-cols-2">
          <AdminFormField label="轮径" required description="后轮独立填写；前后轮轮径可以不同。">
            <Input v-model="specification.rear_wheel.wheel_size" maxlength="32" placeholder="700C / 650B / 29 in" />
          </AdminFormField>
          <AdminFormField label="支持孔数" required description="多个值用逗号分隔，例如 28, 32。">
            <Input v-model="specification.rear_wheel.spoke_hole_counts" placeholder="28, 32" />
          </AdminFormField>
          <AdminFormField label="支持编法" required description="多个值用逗号分隔，例如 2x, 3x。">
            <Input v-model="specification.rear_wheel.lacing_patterns" placeholder="2x, 3x" />
          </AdminFormField>
          <AdminFormField label="BSD（mm）"><Input v-model="specification.rear_wheel.bead_seat_diameter_mm" type="number" min="0.01" step="any" /></AdminFormField>
          <AdminFormField label="材质"><Input v-model="specification.rear_wheel.material" maxlength="120" /></AdminFormField>
          <AdminFormField label="制动接口"><Input v-model="specification.rear_wheel.brake_interface" maxlength="80" placeholder="Center Lock / 6-bolt / rim brake" /></AdminFormField>
          <AdminFormField label="胎圈结构"><Input v-model="specification.rear_wheel.rim_profile" maxlength="80" placeholder="Hooked / Hookless" /></AdminFormField>
          <AdminFormField label="轮胎系统"><Input v-model="specification.rear_wheel.tire_system" maxlength="80" placeholder="Clincher / Tubeless / Tubular" /></AdminFormField>
          <AdminFormField label="框高（mm）"><Input v-model="specification.rear_wheel.rim_depth_mm" type="number" min="0.01" step="any" /></AdminFormField>
          <AdminFormField label="内宽（mm）"><Input v-model="specification.rear_wheel.rim_internal_width_mm" type="number" min="0.01" step="any" /></AdminFormField>
          <AdminFormField label="外宽（mm）"><Input v-model="specification.rear_wheel.rim_external_width_mm" type="number" min="0.01" step="any" /></AdminFormField>
          <AdminFormField label="花鼓接口"><Input v-model="specification.rear_wheel.hub_interface" maxlength="120" /></AdminFormField>
          <AdminFormField label="轴规格"><Input v-model="specification.rear_wheel.axle_standard" maxlength="120" /></AdminFormField>
          <AdminFormField label="塔基"><Input v-model="specification.rear_wheel.freehub_body" maxlength="120" /></AdminFormField>
          <AdminFormField label="辐条类型"><Input v-model="specification.rear_wheel.spoke_type" maxlength="120" /></AdminFormField>
          <AdminFormField label="后轮重量（g）"><Input v-model="specification.rear_wheel.weight_g" type="number" min="0.01" step="any" /></AdminFormField>
          <AdminFormField label="最大胎压（bar）"><Input v-model="specification.rear_wheel.max_tire_pressure_bar" type="number" min="0.01" step="any" /></AdminFormField>
        </fieldset>
      </section>

      <section class="space-y-4 border-t pt-4 lg:col-span-2" aria-labelledby="wheelset-complete-set-heading">
        <div class="border-b pb-2">
          <h3 id="wheelset-complete-set-heading" class="text-sm font-semibold">整组资料</h3>
        </div>
        <fieldset :disabled="locked" class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <AdminFormField label="整组净重（g）"><Input v-model="specification.net_weight_g" type="number" min="0.01" step="any" /></AdminFormField>
          <AdminFormField label="最大系统重量（kg）"><Input v-model="specification.max_system_weight_kg" type="number" min="0.01" step="any" /></AdminFormField>
          <AdminFormField label="认证" description="多个名称用逗号分隔；只填写已核实的正式认证。">
            <Input v-model="specification.certifications" placeholder="例如 UCI approval" />
          </AdminFormField>
        </fieldset>
      </section>
    </div>
    <p v-if="locked" class="mt-3 text-sm text-muted-foreground">产品级规格由源语言商品维护。</p>
    <p v-if="error" role="alert" class="mt-3 text-sm text-destructive">{{ error }}</p>
  </AdminFormSection>
</template>

<script setup lang="ts">
import AdminFormField from '@/components/admin/AdminFormField.vue'
import AdminFormSection from '@/components/admin/AdminFormSection.vue'
import { Input } from '@/components/ui/input'
import type { WheelsetSpecificationForm } from '@/modules/product/wheelsetDedicatedProductSpecificationFormValuesAndValidation'

withDefaults(defineProps<{
  specification: WheelsetSpecificationForm
  locked?: boolean
  error?: string
}>(), {
  locked: false,
  error: '',
})
</script>
