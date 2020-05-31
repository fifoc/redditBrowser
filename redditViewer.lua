--[[
    RedditViewer - Browse reddit from within OpenComputers!
    Copyright (C) 2020 Rph    

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as
    published by the Free Software Foundation, either version 3 of the
    License, or (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

]]--

local libfif = require('libfif')

local component = require('component')
local server = "https://web.sascha-t.de/fife/getfif?frog="
local fs = require('filesystem')


print("Checking for existence of hardware...")

local gpuExistence = component.isAvailable("gpu")
local internetExistence = component.isAvailable("internet")

if gpuExistence == false or internetExistence == false then
    print("Didn't find all necessary hardware, aborting.")
    return
end

local gpu = component.gpu
local internet = require('internet')


local w,h = gpu.maxResolution()

if w < 160 or h < 50 then
    print("This program requires a T3 GPU to run.")
    return
end

print("Type in the subreddit or username you wish to browse. (without leading /) (Example: r/dankmemes)")
local targetsub = io.read()

local realtargetsub = string.match(targetsub, "r%/[%a%d][%a%d%-%_]+")
if realtargetsub == nil then
    realtargetsub = string.match(targetsub, "u/[%a%d][%a%d%-%_]+")
    if realtargetsub == nil then
        print("invalid name for subreddit")
        return
    end
end

if #realtargetsub > 23 then
    print("Subreddit name too long?")
    return
end


--[[
    UTILITY FUNCTIONS FOR MANAGING THE IMAGE CACHE
]]--
local knownImages = {}
local currentImage = 1
local lastAfter = ""

local function fetch( url )
    local handle = internet.request( url, nil, { -- Faking UA otherwise reddit blocks us rather quickly.
        ["User-Agent"]="Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/81.0.4044.138 Safari/537.36 MozarellaFirefox/1.0"
    } )
    local buffer = ""
    for chunk in handle do buffer = buffer .. chunk end
    return buffer
end

if fs.exists(require('shell').getWorkingDirectory() .. "/json.lua") == false then
	print("Downloading JSON library...")
	local jsondata = fetch("https://raw.githubusercontent.com/rxi/json.lua/master/json.lua")
	local handle = fs.open(require('shell').getWorkingDirectory() .. "/json.lua", "w")
	handle:write(jsondata)
	handle:close()
end

local json = require('json')
local function fetchAdditionalImages()
    gpu.setBackground(0x0)
    gpu.setForeground(0xFFFFFF)
    gpu.set(1,2,"LOADING IMAGE LISTING")
    local attempts = 0
    while true do
        attempts = attempts + 1
        local availableImages = #knownImages - currentImage
        if availableImages > 2 then
            break
        end
        if attempts > 6 then
            error("No images have been served to us since the last 5 request. Abandoning project.")
        end

        -- Make request to reddit
        local resp = fetch("https://reddit.com/" .. realtargetsub .. ".json?limit=10&after=" .. lastAfter)

        resp = json.decode(resp)
        lastAfter = resp.data.after

        if resp.kind == "Listing" then
            for i = 1, #resp.data.children do
                thing = resp.data.children[i]
                if thing.kind == "t3" then
                    thing = thing.data
                    if thing.post_hint == "image" and thing.domain == "i.redd.it" then
                        local png = string.match(thing.url, "https://i%.redd%.it/[%a%d]+%.png")
                        local jpg = string.match(thing.url, "https://i%.redd%.it/[%a%d]+%.jpg")
                        local url = ""
                        if png ~= nil then
                            url = png
                        end
                        if jpg ~= nil then
                            url = jpg
                        end
                        if #url > 1 then
                            purl = string.match(url, "[%a%d]+%.png")
                            jurl = string.match(url, "[%a%d]+%.jpg")
                            if purl ~= nil then
                                url = purl
                            end
                            if jurl ~= nil then
                                url = jurl
                            end
                            table.insert(knownImages, { ["t"] = thing.title, ["u"] = url})
                        end
                    end
                end
            end
        else
            error("Reddit responded with garbage.")
        end
    end
end
print("Populating initial cache...")
fetchAdditionalImages()
print(#knownImages)

local function display()
    -- Download image from the server
    local url = server .. knownImages[currentImage].u
    fs.remove("/tmp/.reddit.fif")
    local handle = fs.open("/tmp/.reddit.fif", "w")
    gpu.set(1,2,"DOWNLOADING IMAGE              ")
    handle:write(fetch(url))
    handle:close()


    libfif("/tmp/.reddit.fif")

    gpu.setBackground(0x0)
    gpu.setForeground(0xFFFFFF)
    gpu.set(1,1, knownImages[currentImage].t)
    w, h = gpu.getResolution()
    gpu.set(1,h,"Arrows: Change images | S: Save to /home | Q: Quit")
end

local oldw, oldh = gpu.getResolution()
local event = require('event')
display()
while true do
    eventid, _, _, code = event.pull()
    if eventid == "key_down" then
        if code == 203 and currentImage > 1 then
            currentImage = currentImage - 1
            display()
        end
        if code == 205 then
            currentImage = currentImage + 1
            fetchAdditionalImages()
            display()
        end
        if code == 31 then
            fs.makeDirectory("/home/reddit-saved")
            local handle, z = fs.open("/home/reddit-saved/" .. realtargetsub:gsub("/", "-") .. "--" .. knownImages[currentImage].u .. ".fif", "w")
            handle:write(fetch(server .. knownImages[currentImage].u))
            handle:close()
        end
        if code == 16 then
            break
        end
    end
end

gpu.setResolution(oldw, oldh)
gpu.setBackground(0x0)
gpu.setForeground(0xFFFFFF)
gpu.fill(1,1,oldw,oldh, " ")
require('term').setCursor( 1, 1 )

