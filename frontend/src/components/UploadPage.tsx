import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"

function CardWithForm() {
  return (
    <Card className="w-[50%]">
      <CardHeader>
        <CardTitle>Upload</CardTitle>
        <CardDescription>Send files to the server</CardDescription>
      </CardHeader>
      <CardContent>
        <form>
          <div className="grid w-full items-center gap-4">
            <div className="flex flex-col space-y-1.5">
              <Label className="cursor-pointer" htmlFor="files">Files</Label>
              <Input id="files" type="file" multiple className="cursor-pointer" />
            </div>
            {/*
            <div className="flex flex-col space-y-1.5">
              <Label htmlFor="framework">Framework</Label>
              <Select>
                <SelectTrigger id="framework">
                  <SelectValue placeholder="Select" />
                </SelectTrigger>
                <SelectContent position="popper">
                  <SelectItem value="next">Next.js</SelectItem>
                  <SelectItem value="sveltekit">SvelteKit</SelectItem>
                  <SelectItem value="astro">Astro</SelectItem>
                  <SelectItem value="nuxt">Nuxt.js</SelectItem>
                </SelectContent>
              </Select>
            </div>
              */
            }
          </div>
        </form>
      </CardContent>
      <CardFooter className="flex justify-between">
        <Button variant="secondary">Clear</Button>
        <Button>Upload</Button>
      </CardFooter>
    </Card>
  )
}

export default CardWithForm
