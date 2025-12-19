import { createEffect, createEvent, createStore, sample } from 'effector'
import { routes } from '@shared/router'
import { routesApi, type CreateRouteRequest, type RoutePoint } from '@features/routes'

export interface RouteForm {
  title: string
  description: string
  city: string
  country: string
  is_public: boolean
  is_linear: boolean
  tags: string[]
  points: RoutePoint[]
}

const initialForm: RouteForm = {
  title: '',
  description: '',
  city: '',
  country: '',
  is_public: true,
  is_linear: true,
  tags: [],
  points: [],
}

export const $form = createStore<RouteForm>(initialForm)

export const titleChanged = createEvent<string>()
export const descriptionChanged = createEvent<string>()
export const cityChanged = createEvent<string>()
export const countryChanged = createEvent<string>()
export const privacyToggled = createEvent()
export const linearToggled = createEvent()
export const tagAdded = createEvent<string>()
export const tagRemoved = createEvent<string>()
export const pointAdded = createEvent<Omit<RoutePoint, 'id' | 'order_index'>>()
export const pointRemoved = createEvent<number>()
export const pointUpdated = createEvent<{ index: number; point: Partial<RoutePoint> }>()
export const formReset = createEvent()

$form
  .on(titleChanged, (state, title) => ({ ...state, title }))
  .on(descriptionChanged, (state, description) => ({ ...state, description }))
  .on(cityChanged, (state, city) => ({ ...state, city }))
  .on(countryChanged, (state, country) => ({ ...state, country }))
  .on(privacyToggled, state => ({ ...state, is_public: !state.is_public }))
  .on(linearToggled, state => ({ ...state, is_linear: !state.is_linear }))
  .on(tagAdded, (state, tag) => ({
    ...state,
    tags: state.tags.includes(tag) ? state.tags : [...state.tags, tag],
  }))
  .on(tagRemoved, (state, tag) => ({
    ...state,
    tags: state.tags.filter(t => t !== tag),
  }))
  .on(pointAdded, (state, point) => ({
    ...state,
    points: [
      ...state.points,
      {
        ...point,
        order_index: state.points.length,
      },
    ],
  }))
  .on(pointRemoved, (state, index) => ({
    ...state,
    points: state.points.filter((_, i) => i !== index).map((p, i) => ({ ...p, order_index: i })),
  }))
  .on(pointUpdated, (state, { index, point }) => ({
    ...state,
    points: state.points.map((p, i) => (i === index ? { ...p, ...point } : p)),
  }))
  .reset(formReset)

export const formSubmitted = createEvent()

export const createRouteFx = createEffect<RouteForm, void>(async form => {
  const request: CreateRouteRequest = {
    title: form.title,
    description: form.description || undefined,
    city: form.city || undefined,
    country: form.country || undefined,
    is_public: form.is_public,
    is_linear: form.is_linear,
    tags: form.tags.length > 0 ? form.tags : undefined,
    points: form.points.map(point => {
      // eslint-disable-next-line @typescript-eslint/no-unused-vars
      const { id, order_index, ...rest } = point
      return rest
    }),
  }

  await routesApi.createRoute(request)
})

sample({
  clock: formSubmitted,
  source: $form,
  target: createRouteFx,
})

sample({
  clock: createRouteFx.done,
  target: routes.home.navigate.prepend(() => ({ params: {}, query: {} })),
})
