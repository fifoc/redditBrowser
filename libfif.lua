--[[
    Libfif - Reference fif renderer
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

local gpu = require('component').gpu
--gpu.setBackground(0xFFFFFF)
local gpu_set, gpu_setbg, gpu_setfg, gpu_fill = gpu.set, gpu.setBackground, gpu.setForeground, gpu.fill
local unicode_char = require('unicode').char
local string_sub, string_byte = string.sub, string.byte

return function(path)
local handle = io.open(path, "rb")

handle:read(6)
gpu.setResolution(string_byte(handle:read(1)), string_byte(handle:read(1)))

local data = ""
local s = ""
local readexe = true
while readexe == true do
  s = handle:read(1024)
  if s ~= nil then
    data = data .. s
  else
    readexe = false
  end
end
local offset = 1
local execution = true
while execution == true do
  local instruction = string_byte(string_sub(data, offset, offset), 1)
  offset = offset + 1
 -- print(instruction)
  if instruction == 0x01 then
    local data = string_sub(data, offset, offset + 2)
    offset = offset + 3
    local r = string_byte(data, 1)
    local g = string_byte(data, 2)
    local b = string_byte(data, 3)
    local color = r * 0x10000 + g * 0x100 + b
    gpu_setbg(color)
    
  end
  if instruction == 0x02 then
    local data = string_sub(data, offset, offset + 2)
    offset = offset + 3
    local r = string_byte(data, 1)
    local g = string_byte(data, 2)
    local b = string_byte(data, 3)
    local color = r * 0x10000 + g * 0x100 + b
    gpu_setfg(color)
  end
  if instruction == 0x10 then
    local cords = string_sub(data, offset, offset + 1)
    offset = offset + 2
    local x = string_byte(cords, 1) + 1
    local y = string_byte(cords, 2) + 1
    local size = string_byte(string_sub(data, offset, offset))
    offset = offset + 1
    local pixelInfo = string_sub(data, offset, offset + size - 1)
    offset = offset + size
    local gpustring = ""
    for i=1, #pixelInfo do
      local char = string_byte(string_sub(pixelInfo, i, i))
      char = char + 0x2800
      gpustring = gpustring .. unicode_char(char)
    end
    gpu_set(x, y, gpustring)
  end
  if instruction == 0x11 then
    local cords = string_sub(data, offset, offset + 4)
    offset = offset + 5
    local x = string_byte(cords, 1) + 1
    local y = string_byte(cords, 2) + 1
    local w = string_byte(cords, 3)
    local h = string_byte(cords, 4)
    local char = string_byte(cords, 5)
    char = unicode_char(0x2800 + char)
    gpu_fill(x, y, w, h, char)
  end
  if instruction == 0x12 then
    local off = string_sub(data, offset, offset)
    offset = offset + 1
    off = string_byte(off)
    os.sleep(off / 100)
  end
  if instruction == 0x13 then
    local cords = string_sub(data, offset ,offset + 1)
      offset = offset + 2
    local x = string_byte(cords, 1) + 1
    local y = string_byte(cords, 2) + 1
    local size = string_byte(string_sub(data,offset,offset))
    offset = offset + 1
    local pixelInfo = string_sub(data, offset, offset + size - 1)
    offset = offset + size
    local gpustring = ""
    for i=1,#pixelInfo do
      local char = string_byte(string_sub(pixelInfo, i, i))
      char = char + 0x2800
      gpustring = gpustring .. unicode_char(char)
    end
    gpu.set(x, y, gpustring, true)
  end
  if instruction == 0x20 then
    execution = false
  end
--  os.sleep(0.05)
end

  handle:close()
end
