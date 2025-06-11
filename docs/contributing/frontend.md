# Contributing to Nixopus Frontend

This guide provides detailed instructions for contributing to the Nixopus frontend codebase.

## Setup for Frontend Development

1. **Prerequisites**
   - Node.js 18.0 or higher
   - Yarn package manager
   - A running instance of the Nixopus API (locally or remotely)

2. **Environment Setup**

   ```bash
   # Clone the repository
   git clone https://github.com/raghavyuva/nixopus.git
   cd nixopus
   
   # Copy environment template
   cp view/.env.sample view/.env.local
   
   # Install dependencies
   cd view
   yarn install
   ```

3. **Running the Development Server**

   ```bash
   cd view
   yarn dev
   ```

   The development server will start on `http://localhost:3000` by default.

## Project Structure

The frontend follows Next.js App Router structure with the following organization:

```
view/
├── app/               # Next.js App Router pages
│   ├── layout.tsx     # Root layout
│   ├── page.tsx       # Home page
│   ├── dashboard/     # Dashboard pages
│   ├── login/         # Authentication pages
│   ├── settings/      # Settings pages
│   └── ...
├── components/        # React components
│   ├── ui/            # UI components
│   ├── layout/        # Layout components
│   └── features/      # Feature-specific components
├── hooks/             # Custom React hooks
├── lib/               # Utility functions
├── redux/             # Redux store setup
└── types/             # TypeScript type definitions
```

## Adding a New Feature

1. **Create a New Branch**

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Plan Your Implementation**

   Consider where your feature fits within the application:
   - Is it a new page/route?
   - Is it a component enhancement?
   - Does it modify existing behavior?

3. **Implementation Steps**

   ### For a New Page

   Create a new directory in the appropriate place in the App Router structure:

   ```
   view/app/your-feature/
   ├── page.tsx        # Main page component
   ├── layout.tsx      # Optional layout
   └── components/     # Page-specific components
   ```

   Example `page.tsx`:

   ```tsx
   'use client'
   
   import { useEffect, useState } from 'react'
   import { useDispatch, useSelector } from 'react-redux'
   import { fetchYourFeatureData } from '@/redux/your-feature/actions'
   import YourFeatureComponent from './components/YourFeatureComponent'
   
   export default function YourFeaturePage() {
     const dispatch = useDispatch()
     const { data, loading, error } = useSelector(state => state.yourFeature)
     
     useEffect(() => {
       dispatch(fetchYourFeatureData())
     }, [dispatch])
     
     if (loading) return <div>Loading...</div>
     if (error) return <div>Error: {error}</div>
     
     return (
       <div className="container mx-auto py-8">
         <h1 className="text-2xl font-bold mb-4">Your Feature</h1>
         <YourFeatureComponent data={data} />
       </div>
     )
   }
   ```

   ### For a New Component

   Create components in the appropriate directory:

   ```tsx
   // view/components/features/your-feature/YourComponent.tsx
   
   import { useState } from 'react'
   import { Button } from '@/components/ui/button'
   
   interface YourComponentProps {
     title: string
     onAction: () => void
   }
   
   export default function YourComponent({ title, onAction }: YourComponentProps) {
     const [isActive, setIsActive] = useState(false)
     
     return (
       <div className="border rounded-lg p-4">
         <h2 className="text-xl font-semibold">{title}</h2>
         <p className="text-gray-600 my-2">Your component description</p>
         <Button 
           variant={isActive ? 'default' : 'outline'} 
           onClick={() => {
             setIsActive(!isActive)
             onAction()
           }}
         >
           {isActive ? 'Active' : 'Inactive'}
         </Button>
       </div>
     )
   }
   ```

   ### For Redux Integration

   Create the necessary Redux files:

   ```
   view/redux/your-feature/
   ├── slice.ts        # Redux Toolkit slice
   ├── actions.ts      # Async actions
   └── selectors.ts    # Selectors
   ```

   To write down API calls, create a new file in the `services` directory using Redux Toolkit Query:

   Example:

   ```tsx
   // view/redux/services/<featureFolder>/<featureName>Api.ts
   import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react'
   import { baseQueryWithReauth } from '@/redux/base-query'
   import { 
     <EntityType>, 
     <CreateDto>, 
     <UpdateDto> 
   } from '@/redux/types/<featureFolder>'
   import { API_ENDPOINTS } from '@/redux/api-conf'

   export const <featureName>Api = createApi({
     reducerPath: '<featureName>Api',
     baseQuery: baseQueryWithReauth,
     tagTypes: ['<TagType>'],
     endpoints: (builder) => ({
       get<FeaturePascalPlural>: builder.query<<EntityType>[], void>({
         query: () => ({ url: API_ENDPOINTS.<FEATURE_PLURAL>, method: 'GET' }),
         transformResponse: (res: { data: <EntityType>[] }) => res.data,
         providesTags: (result) =>
           result
             ? [
                 ...result.map(({ id }) => ({ type: '<TagType>' as const, id })),
                 { type: '<TagType>', id: 'LIST' },
               ]
             : [{ type: '<TagType>', id: 'LIST' }],
       }),
       create<FeaturePascal>: builder.mutation<<EntityType>, <CreateDto>>({
         query: (body) => ({
           url: API_ENDPOINTS.ADD_<FEATURE_UPPER>,
           method: 'POST',
           body,
         }),
         invalidatesTags: [{ type: '<TagType>', id: 'LIST' }],
       }),
       update<FeaturePascal>: builder.mutation<<EntityType>, <UpdateDto>>({
         query: (body) => ({
           url: API_ENDPOINTS.UPDATE_<FEATURE_UPPER>,
           method: 'PUT',
           body,
         }),
         invalidatesTags: [{ type: '<TagType>', id: 'LIST' }],
       }),
       delete<FeaturePascal>: builder.mutation<void, string>({
         query: (id) => ({
           url: API_ENDPOINTS.DELETE_<FEATURE_UPPER>,
           method: 'DELETE',
           body: { id },
         }),
         invalidatesTags: [{ type: '<TagType>', id: 'LIST' }],
       }),
     }),
   })

   export const {
     useGet<FeaturePascalPlural>Query,
     useCreate<FeaturePascal>Mutation,
     useUpdate<FeaturePascal>Mutation,
     useDelete<FeaturePascal>Mutation,
   } = <featureName>Api
   ```

4. **Add Tests**

   Create tests using Jest and React Testing Library:

   ```tsx
   // view/components/features/your-feature/YourComponent.test.tsx
   import { render, screen, fireEvent } from '@testing-library/react'
   import YourComponent from './YourComponent'
   
   describe('YourComponent', () => {
     const mockOnAction = jest.fn()
     
     beforeEach(() => {
       jest.clearAllMocks()
     })
     
     it('renders correctly with props', () => {
       render(<YourComponent title="Test Title" onAction={mockOnAction} />)
       expect(screen.getByText('Test Title')).toBeInTheDocument()
       expect(screen.getByText('Your component description')).toBeInTheDocument()
       expect(screen.getByRole('button')).toHaveTextContent('Inactive')
     })
     
     it('calls onAction when button is clicked', () => {
       render(<YourComponent title="Test Title" onAction={mockOnAction} />)
       fireEvent.click(screen.getByRole('button'))
       expect(mockOnAction).toHaveBeenCalledTimes(1)
       expect(screen.getByRole('button')).toHaveTextContent('Active')
     })
   })
   ```

## UI Guidelines

1. **Use Existing UI Components**

   Nixopus uses Shadcn UI components based on Radix UI for consistency. Explore the `components/ui` directory before creating new components.

2. **Follow Design System**

   - Use the project's color palette defined in `tailwind.config.js`
   - Follow spacing and typography guidelines
   - Maintain responsive designs that work on all screen sizes

3. **Accessibility**

   - Use semantic HTML elements
   - Add proper ARIA attributes when needed
   - Ensure keyboard navigation works
   - Maintain sufficient color contrast

4. **Component Organization**

   - Break down large components into smaller, reusable pieces
   - Use composition with children props when appropriate
   - Follow the container/presentation component pattern

## Code Style and Guidelines

1. **TypeScript**

   - Use proper type definitions
   - Avoid `any` types when possible
   - Create interfaces for component props
   - Use type guards for conditional rendering

2. **React Best Practices**

   - Use functional components with hooks
   - Memoize expensive calculations with useMemo
   - Optimize renders with useCallback for handler functions
   - Use React.memo for pure components that render often

3. **CSS and Styling**

   - Use Tailwind CSS for styling
   - Follow the project's class naming conventions
   - Group related classes for better readability

4. **Code Formatting**

   The project uses ESLint and Prettier for code formatting:

   ```bash
   # Check for lint errors
   yarn lint
   
   # Fix lint errors
   yarn lint:fix
   ```

## Testing Your Changes

1. **Run the Development Server**

   ```bash
   yarn dev
   ```

2. **Run Tests**

   ```bash
   yarn test
   yarn test:watch  # Watch mode
   ```

3. **Build for Production**

   ```bash
   yarn build
   ```

## Common Pitfalls

1. Not updating types when modifying data structures
2. Forgetting to handle loading and error states
3. Not considering mobile responsiveness
4. Creating duplicate functionality instead of reusing components
5. Not testing edge cases and error scenarios

## Submitting Your Contribution

1. **Commit Changes**

   ```bash
   git add .
   git commit -m "feat: add your feature"
   ```

2. **Push and Create a Pull Request**

   ```bash
   git push origin feature/your-feature-name
   ```

3. Follow the PR template and provide detailed information about your changes.

## Need Help?

If you need assistance, feel free to:

- Create an issue on GitHub
- Reach out on the project's Discord channel
- Contact the maintainers directly

Thank you for contributing to Nixopus!
